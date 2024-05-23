package lockfile

type PackageJsonMatcher struct{}

func (m PackageJsonMatcher) GetSourceFile(lockfile DepFile) (DepFile, error) {
	return lockfile.Open("package.json")
}

func (m PackageJsonMatcher) Match(_ DepFile, _ []PackageDetails) error {
	return nil
}

var _ Matcher = PackageJsonMatcher{}
