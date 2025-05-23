package lockfile

import (
	"log"
	"path/filepath"

	"github.com/google/osv-scanner/pkg/models"
)

const gemfileFilename = "Gemfile"

// Source: https://www.bundler.cn/guides/groups.html
var knownBundlerDevelopmentGroups = map[string]struct{}{
	"dev":         {},
	"development": {},
	"test":        {},
	"ci":          {},
	"cucumber":    {},
	"linting":     {},
	"rubocop":     {},
}

type gemMetadata struct {
	name          string
	groups        []string
	blockLine     models.Position
	blockColumn   models.Position
	nameLine      models.Position
	nameColumn    models.Position
	versionLine   *models.Position
	versionColumn *models.Position
}

type GemfileMatcher struct{}

func (matcher GemfileMatcher) GetSourceFile(lockfile DepFile) (DepFile, error) {
	lockfileDir := filepath.Dir(lockfile.Path())
	sourceFilePath := filepath.Join(lockfileDir, gemfileFilename)
	file, err := OpenLocalDepFile(sourceFilePath)

	return file, err
}

func (matcher GemfileMatcher) Match(sourceFile DepFile, packages []PackageDetails) error {
	packagesByName := indexPackages(packages)

	treeResult, err := ParseFile(sourceFile, Ruby)
	if err != nil {
		return err
	}
	defer treeResult.Close()

	rootGems, err := findGems(treeResult.Node)
	if err != nil {
		return err
	}
	enrichPackagesWithLocation(sourceFile, rootGems, packagesByName)

	remainingGems, err := findGroupedGems(treeResult.Node)
	if err != nil {
		return err
	}
	enrichPackagesWithLocation(sourceFile, remainingGems, packagesByName)

	return nil
}

func findGems(node *Node) ([]gemMetadata, error) {
	// Matches method calls to `gem`
	// extracting the gem dependency name and gem dependency requirement
	gemQueryString := `(
		(call
			method: (identifier) @method_name
			(#match? @method_name "gem")
			arguments: (argument_list
				.
				(comment)*
				.
				(string) @gem_name
				.
				(comment)*
				.
				(string)? @gem_requirement
				.
				(_)*
				.
			)
		) @gem_call
	)`

	gems := make([]gemMetadata, 0)
	err := node.Query(gemQueryString, func(match *MatchResult) error {
		callNode := match.FindFirstByName("gem_call")

		dependencyNameNode := match.FindFirstByName("gem_name")
		dependencyName, err := node.Ctx.ExtractTextValue(dependencyNameNode.TSNode)
		if err != nil {
			return err
		}

		requirementNode := match.FindFirstByName("gem_requirement")

		groups, err := findGroupsInPairs(callNode)
		if err != nil {
			return err
		}

		metadata := gemMetadata{
			name:        dependencyName,
			groups:      groups,
			blockLine:   models.Position{Start: int(callNode.TSNode.StartPosition().Row) + 1, End: int(callNode.TSNode.EndPosition().Row) + 1},
			blockColumn: models.Position{Start: int(callNode.TSNode.StartPosition().Column) + 1, End: int(callNode.TSNode.EndPosition().Column) + 1},
			nameLine:    models.Position{Start: int(dependencyNameNode.TSNode.StartPosition().Row) + 1, End: int(dependencyNameNode.TSNode.EndPosition().Row) + 1},
			nameColumn:  models.Position{Start: int(dependencyNameNode.TSNode.StartPosition().Column) + 1, End: int(dependencyNameNode.TSNode.EndPosition().Column) + 1},
		}

		if requirementNode != nil {
			metadata.versionLine = &models.Position{Start: int(requirementNode.TSNode.StartPosition().Row) + 1, End: int(requirementNode.TSNode.EndPosition().Row) + 1}
			metadata.versionColumn = &models.Position{Start: int(requirementNode.TSNode.StartPosition().Column) + 1, End: int(requirementNode.TSNode.EndPosition().Column) + 1}
		}

		gems = append(gems, metadata)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return gems, nil
}

func findGroupedGems(node *Node) ([]gemMetadata, error) {
	// Matches method calls to `group` with a block
	// extracting the groups and the block (which will contain the calls to `gem`)
	groupQueryString := `(
		(call
			method: (identifier) @method_name
			(#match? @method_name "group")
			arguments: (argument_list
				.
				[
					(simple_symbol)
					(string)
					(comment)
					","
				]*
				.
			) @group_keys
			block: (_) @block
		)
	)`

	gems := make([]gemMetadata, 0)
	err := node.Query(groupQueryString, func(match *MatchResult) error {
		groupKeysNode := match.FindFirstByName("group_keys")
		groups, err := node.Ctx.ExtractTextValues(groupKeysNode.TSNode)
		if err != nil {
			return err
		}

		blockNode := match.FindFirstByName("block")
		blockGems, err := findGems(blockNode)
		if err != nil {
			return err
		}

		// Top-level group always applies to all gem defined groups
		for idx := range blockGems {
			blockGems[idx].groups = groups
		}

		gems = append(gems, blockGems...)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return gems, nil
}

func findGroupsInPairs(node *Node) ([]string, error) {
	// Matches pairs of key-value where the key is "group"
	// This can be a simple pair or a pair used inside other structures like a hash
	pairQuery := `(
		(pair
			key: [(hash_key_symbol) (simple_symbol)] @pair_key
			(#match? @pair_key "group")
			value: [(array) (simple_symbol) (string)] @pair_value
		)
	)`

	var groups []string
	err := node.Query(pairQuery, func(match *MatchResult) error {
		pairValueNode := match.FindFirstByName("pair_value")
		pairGroups, err := node.Ctx.ExtractTextValues(pairValueNode.TSNode)
		if err != nil {
			return err
		}

		groups = append(groups, pairGroups...)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return groups, nil
}

func indexPackages(packages []PackageDetails) map[string]*PackageDetails {
	result := make(map[string]*PackageDetails)
	for index, pkg := range packages {
		result[pkg.Name] = &packages[index]
	}

	return result
}

func enrichPackagesWithLocation(sourceFile DepFile, gems []gemMetadata, packagesByName map[string]*PackageDetails) {
	for _, gem := range gems {
		pkg, ok := packagesByName[gem.name]
		// If packages exist in the Gemfile but not in the Gemfile.lock, we skip the package as we treat the lockfile as
		// the source of truth
		if !ok {
			log.Printf("Skipping package %q from Gemfile as it does not exist in the Gemfile.lock\n", gem.name)
			continue
		}

		pkg.BlockLocation = models.FilePosition{
			Line:     gem.blockLine,
			Column:   gem.blockColumn,
			Filename: sourceFile.Path(),
		}
		pkg.NameLocation = &models.FilePosition{
			Line:     gem.nameLine,
			Column:   gem.nameColumn,
			Filename: sourceFile.Path(),
		}
		if gem.versionLine != nil && gem.versionColumn != nil {
			pkg.VersionLocation = &models.FilePosition{
				Line:     *gem.versionLine,
				Column:   *gem.versionColumn,
				Filename: sourceFile.Path(),
			}
		}
		if len(gem.groups) > 0 {
			pkg.DepGroups = gem.groups
		}
	}
}
