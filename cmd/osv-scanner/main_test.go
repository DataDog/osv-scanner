// main cannot be accessed directly, so cannot use main_test
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/google/osv-scanner/pkg/models"

	sbom_test "github.com/google/osv-scanner/internal/utility/sbom"

	"github.com/CycloneDX/cyclonedx-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/google/osv-scanner/internal/testutility"
	"github.com/urfave/cli/v2"
)

type cliTestCase struct {
	name string
	args []string
	exit int
}

type locationTestCase struct {
	name          string
	args          []string
	wantExitCode  int
	wantFilePaths []string
}

type encodingTestCase struct {
	encoding string
}

func runCli(t *testing.T, tc cliTestCase) (string, string) {
	t.Helper()

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	ec := run(tc.args, stdout, stderr)

	if ec != tc.exit {
		t.Errorf("cli exited with code %d, not %d", ec, tc.exit)
	}

	return testutility.NormalizeStdStream(t, stdout), testutility.NormalizeStdStream(t, stderr)
}

func testCli(t *testing.T, tc cliTestCase) {
	t.Helper()

	stdout, stderr := runCli(t, tc)

	testutility.NewSnapshot().MatchText(t, stdout)
	testutility.NewSnapshot().MatchText(t, stderr)
}

func TestRun(t *testing.T) {
	t.Parallel()

	tests := []cliTestCase{
		{
			name: "",
			args: []string{""},
			exit: 0,
		},
		{
			name: "",
			args: []string{"", "--version"},
			exit: 0,
		},
		// one specific supported lockfile
		{
			name: "one specific supported lockfile",
			args: []string{"", "./fixtures/locks-many/composer.lock"},
			exit: 0,
		},
		// one specific supported sbom with vulns
		{
			name: "folder of supported sbom with vulns",
			args: []string{"", "--config=./fixtures/osv-scanner-empty-config.toml", "./fixtures/sbom-insecure/"},
			exit: 0,
		},
		// one specific supported sbom with vulns
		{
			name: "one specific supported sbom with vulns",
			args: []string{"", "--config=./fixtures/osv-scanner-empty-config.toml", "--sbom", "./fixtures/sbom-insecure/alpine.cdx.xml"},
			exit: 0,
		},
		// one specific unsupported lockfile
		{
			name: "",
			args: []string{"", "./fixtures/locks-many/not-a-lockfile.toml"},
			exit: 0,
		},
		// all supported lockfiles in the directory should be checked
		{
			name: "Scan locks-many",
			args: []string{"", "./fixtures/locks-many"},
			exit: 0,
		},
		// all supported lockfiles in the directory should be checked
		{
			name: "all supported lockfiles in the directory should be checked",
			args: []string{"", "./fixtures/locks-many-with-invalid"},
			exit: 0,
		},
		// only the files in the given directories are checked by default (no recursion)
		{
			name: "only the files in the given directories are checked by default (no recursion)",
			args: []string{"", "./fixtures/locks-one-with-nested"},
			exit: 0,
		},
		// nested directories are checked when `--recursive` is passed
		{
			name: "nested directories are checked when `--recursive` is passed",
			args: []string{"", "--recursive", "./fixtures/locks-one-with-nested"},
			exit: 0,
		},
		// .gitignored files
		{
			name: "",
			args: []string{"", "--recursive", "./fixtures/locks-gitignore"},
			exit: 0,
		},
		// ignoring .gitignore
		{
			name: "",
			args: []string{"", "--recursive", "--no-ignore", "./fixtures/locks-gitignore"},
			exit: 0,
		},
		// output with json
		{
			name: "json output 1",
			args: []string{"", "--json", "./fixtures/locks-many/composer.lock"},
			exit: 0,
		},
		{
			name: "json output 2",
			args: []string{"", "--format", "json", "./fixtures/locks-many/composer.lock"},
			exit: 0,
		},
		// output format: sarif
		{
			name: "Empty sarif output",
			args: []string{"", "--format", "sarif", "./fixtures/locks-many/composer.lock"},
			exit: 0,
		},
		{
			name: "Sarif with vulns",
			args: []string{"", "--format", "sarif", "--config", "./fixtures/osv-scanner-empty-config.toml", "./fixtures/locks-many/package-lock.json"},
			exit: 0,
		},
		// output format: gh-annotations
		{
			name: "Empty gh-annotations output",
			args: []string{"", "--format", "gh-annotations", "./fixtures/locks-many/composer.lock"},
			exit: 0,
		},
		{
			name: "gh-annotations with vulns",
			args: []string{"", "--format", "gh-annotations", "--config", "./fixtures/osv-scanner-empty-config.toml", "./fixtures/locks-many/package-lock.json"},
			exit: 0,
		},
		// output format: markdown table
		{
			name: "",
			args: []string{"", "--format", "markdown", "--config", "./fixtures/osv-scanner-empty-config.toml", "./fixtures/locks-many/package-lock.json"},
			exit: 0,
		},
		// output format: unsupported
		{
			name: "",
			args: []string{"", "--format", "unknown", "./fixtures/locks-many/composer.lock"},
			exit: 127,
		},
		// one specific supported lockfile with ignore
		{
			name: "one specific supported lockfile with ignore",
			args: []string{"", "./fixtures/locks-test-ignore/package-lock.json"},
			exit: 0,
		},
		{
			name: "invalid --verbosity value",
			args: []string{"", "--verbosity", "unknown", "./fixtures/locks-many/composer.lock"},
			exit: 127,
		},
		{
			name: "verbosity level = error",
			args: []string{"", "--verbosity", "error", "--format", "table", "./fixtures/locks-many/composer.lock"},
			exit: 0,
		},
		{
			name: "verbosity level = info",
			args: []string{"", "--verbosity", "info", "--format", "table", "./fixtures/locks-many/composer.lock"},
			exit: 0,
		},
		// Go project with an overridden go version
		{
			name: "Go project with an overridden go version",
			args: []string{"", "./fixtures/go-project"},
			exit: 0,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			testCli(t, tt)
		})
	}
}

func TestRunCallAnalysis(t *testing.T) {
	t.Parallel()

	// Switch to acceptance test if this takes too long, or when we add rust tests
	// testutility.SkipIfNotAcceptanceTesting(t, "Takes a while to run")

	tests := []cliTestCase{
		{
			name: "Run with govulncheck",
			args: []string{"",
				"--call-analysis=go",
				"--config=./fixtures/osv-scanner-empty-config.toml",
				"./fixtures/call-analysis-go-project"},
			exit: 0,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			testCli(t, tt)
		})
	}
}

func TestRun_LockfileWithExplicitParseAs(t *testing.T) {
	t.Parallel()

	tests := []cliTestCase{
		// unsupported parse-as
		{
			name: "",
			args: []string{"", "-L", "my-file:./fixtures/locks-many/composer.lock"},
			exit: 127,
		},
		// empty is default
		{
			name: "",
			args: []string{
				"",
				"-L",
				":" + filepath.FromSlash("./fixtures/locks-many/composer.lock"),
			},
			exit: 0,
		},
		// empty works as an escape (no fixture because it's not valid on Windows)
		{
			name: "",
			args: []string{
				"",
				"-L",
				":" + filepath.FromSlash("./path/to/my:file"),
			},
			exit: 127,
		},
		{
			name: "",
			args: []string{
				"",
				"-L",
				":" + filepath.FromSlash("./path/to/my:project/package-lock.json"),
			},
			exit: 127,
		},
		// one lockfile with local path
		{
			name: "one lockfile with local path",
			args: []string{"", "--lockfile=go.mod:./fixtures/locks-many/replace-local.mod"},
			exit: 0,
		},
		// when an explicit parse-as is given, it's applied to that file
		{
			name: "",
			args: []string{
				"",
				"-L",
				"package-lock.json:" + filepath.FromSlash("./fixtures/locks-insecure/my-package-lock.json"),
				filepath.FromSlash("./fixtures/locks-insecure"),
			},
			exit: 0,
		},
		// multiple, + output order is deterministic
		{
			name: "",
			args: []string{
				"",
				"-L", "package-lock.json:" + filepath.FromSlash("./fixtures/locks-insecure/my-package-lock.json"),
				"-L", "yarn.lock:" + filepath.FromSlash("./fixtures/locks-insecure/my-yarn.lock"),
				filepath.FromSlash("./fixtures/locks-insecure"),
			},
			exit: 0,
		},
		{
			name: "",
			args: []string{
				"",
				"-L", "yarn.lock:" + filepath.FromSlash("./fixtures/locks-insecure/my-yarn.lock"),
				"-L", "package-lock.json:" + filepath.FromSlash("./fixtures/locks-insecure/my-package-lock.json"),
				filepath.FromSlash("./fixtures/locks-insecure"),
			},
			exit: 0,
		},
		// files that error on parsing stop parsable files from being checked
		{
			name: "",
			args: []string{
				"",
				"-L",
				"Cargo.lock:" + filepath.FromSlash("./fixtures/locks-insecure/my-package-lock.json"),
				filepath.FromSlash("./fixtures/locks-insecure"),
				filepath.FromSlash("./fixtures/locks-many"),
			},
			exit: 127,
		},
		// parse-as takes priority, even if it's wrong
		{
			name: "",
			args: []string{
				"",
				"-L",
				"package-lock.json:" + filepath.FromSlash("./fixtures/locks-many/yarn.lock"),
			},
			exit: 127,
		},
		// "apk-installed" is supported
		{
			name: "",
			args: []string{
				"",
				"-L",
				"apk-installed:" + filepath.FromSlash("./fixtures/locks-many/installed"),
			},
			exit: 0,
		},
		// "dpkg-status" is supported
		{
			name: "",
			args: []string{
				"",
				"-L",
				"dpkg-status:" + filepath.FromSlash("./fixtures/locks-many/status"),
			},
			exit: 0,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			testCli(t, tt)
		})
	}
}

// TestRun_GithubActions tests common actions the github actions reusable workflow will run
func TestRun_GithubActions(t *testing.T) {
	t.Parallel()

	tests := []cliTestCase{
		{
			name: "scanning osv-scanner custom format",
			args: []string{"", "-L", "osv-scanner:./fixtures/locks-insecure/osv-scanner-flutter-deps.json"},
			exit: 0,
		},
		{
			name: "scanning osv-scanner custom format output json",
			args: []string{"", "-L", "osv-scanner:./fixtures/locks-insecure/osv-scanner-flutter-deps.json", "--format=sarif"},
			exit: 0,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			testCli(t, tt)
		})
	}
}

func TestRun_LocalDatabases(t *testing.T) {
	t.Parallel()

	tests := []cliTestCase{
		// one specific supported lockfile
		{
			name: "",
			args: []string{"", "--experimental-local-db", "./fixtures/locks-many/composer.lock"},
			exit: 0,
		},
		// one specific supported sbom with vulns
		{
			name: "",
			args: []string{"", "--experimental-local-db", "--config=./fixtures/osv-scanner-empty-config.toml", "./fixtures/sbom-insecure/postgres-stretch.cdx.xml"},
			exit: 0,
		},
		// one specific unsupported lockfile
		{
			name: "",
			args: []string{"", "--experimental-local-db", "./fixtures/locks-many/not-a-lockfile.toml"},
			exit: 0,
		},
		// all supported lockfiles in the directory should be checked
		{
			name: "",
			args: []string{"", "--experimental-local-db", "./fixtures/locks-many"},
			exit: 0,
		},
		// all supported lockfiles in the directory should be checked
		{
			name: "",
			args: []string{"", "--experimental-local-db", "./fixtures/locks-many-with-invalid"},
			exit: 0,
		},
		// only the files in the given directories are checked by default (no recursion)
		{
			name: "",
			args: []string{"", "--experimental-local-db", "./fixtures/locks-one-with-nested"},
			exit: 0,
		},
		// nested directories are checked when `--recursive` is passed
		{
			name: "",
			args: []string{"", "--experimental-local-db", "--recursive", "./fixtures/locks-one-with-nested"},
			exit: 0,
		},
		// .gitignored files
		{
			name: "",
			args: []string{"", "--experimental-local-db", "--recursive", "./fixtures/locks-gitignore"},
			exit: 0,
		},
		// ignoring .gitignore
		{
			name: "",
			args: []string{"", "--experimental-local-db", "--recursive", "--no-ignore", "./fixtures/locks-gitignore"},
			exit: 0,
		},
		// output with json
		{
			name: "",
			args: []string{"", "--experimental-local-db", "--json", "./fixtures/locks-many/composer.lock"},
			exit: 0,
		},
		{
			name: "",
			args: []string{"", "--experimental-local-db", "--format", "json", "./fixtures/locks-many/composer.lock"},
			exit: 0,
		},
		// output format: markdown table
		{
			name: "",
			args: []string{"", "--experimental-local-db", "--format", "markdown", "./fixtures/locks-many/composer.lock"},
			exit: 0,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if testutility.IsAcceptanceTest() {
				testDir := testutility.CreateTestDir(t)
				old := tt.args
				tt.args = []string{"", "--experimental-local-db-path", testDir}
				tt.args = append(tt.args, old[1:]...)
			}

			// run each test twice since they should provide the same output,
			// and the second run should be fast as the db is already available
			testCli(t, tt)
			testCli(t, tt)
		})
	}
}

func TestRun_Licenses(t *testing.T) {
	t.Parallel()
	tests := []cliTestCase{
		{
			name: "No vulnerabilities with license summary",
			args: []string{"", "--experimental-licenses-summary", "./fixtures/locks-many"},
			exit: 0,
		},
		{
			name: "No vulnerabilities with license summary in markdown",
			args: []string{"", "--experimental-licenses-summary", "--format=markdown", "./fixtures/locks-many"},
			exit: 0,
		},
		{
			name: "Vulnerabilities and license summary",
			args: []string{"", "--experimental-licenses-summary", "--config=./fixtures/osv-scanner-empty-config.toml", "./fixtures/locks-many/package-lock.json"},
			exit: 0,
		},
		{
			name: "Vulnerabilities and license violations with allowlist",
			args: []string{"", "--experimental-licenses", "MIT", "--config=./fixtures/osv-scanner-empty-config.toml", "./fixtures/locks-many/package-lock.json"},
			exit: 0,
		},
		{
			name: "Vulnerabilities and all license violations allowlisted",
			args: []string{"", "--experimental-licenses", "Apache-2.0", "--config=./fixtures/osv-scanner-empty-config.toml", "./fixtures/locks-many/package-lock.json"},
			exit: 0,
		},
		{
			name: "Some packages with license violations and show-all-packages in json",
			args: []string{"", "--format=json", "--experimental-licenses", "MIT", "--experimental-all-packages", "./fixtures/locks-licenses/package-lock.json"},
			exit: 0,
		},
		{
			name: "Some packages with license violations in json",
			args: []string{"", "--format=json", "--experimental-licenses", "MIT", "./fixtures/locks-licenses/package-lock.json"},
			exit: 0,
		},
		{
			name: "No license violations and show-all-packages in json",
			args: []string{"", "--format=json", "--experimental-licenses", "MIT,Apache-2.0", "--experimental-all-packages", "./fixtures/locks-licenses/package-lock.json"},
			exit: 0,
		},
		{
			name: "Licenses in summary mode json",
			args: []string{"", "--format=json", "--experimental-licenses-summary", "./fixtures/locks-licenses/package-lock.json"},
			exit: 0,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			testCli(t, tt)
		})
	}
}

func TestRun_WithoutHostPathInformation(t *testing.T) {
	t.Parallel()
	tests := []locationTestCase{
		// one specific supported lockfile
		{
			name:          "one specific supported lockfile",
			args:          []string{"", "--experimental-only-packages", "--format=cyclonedx-1-5", "--consider-scan-path-as-root", "./fixtures/locks-many/yarn.lock"},
			wantExitCode:  0,
			wantFilePaths: []string{"/package.json"},
		},
		{
			name:         "Multiple lockfiles",
			args:         []string{"", "--experimental-only-packages", "--format=cyclonedx-1-5", "--consider-scan-path-as-root", "./fixtures/locks-many"},
			wantExitCode: 0,
			wantFilePaths: []string{
				"/package-lock.json", // TODO: remove when NPM is using the JSON matcher
				"/package.json",
			},
		},
		{
			name:          "one specific supported lockfile (relative path)",
			args:          []string{"", "--experimental-only-packages", "--format=cyclonedx-1-5", "--paths-relative-to-scan-dir", "./fixtures/locks-many/yarn.lock"},
			wantExitCode:  0,
			wantFilePaths: []string{"package.json"},
		},
		{
			name:         "Multiple lockfiles (relative path)",
			args:         []string{"", "--experimental-only-packages", "--format=cyclonedx-1-5", "--paths-relative-to-scan-dir", "./fixtures/locks-many"},
			wantExitCode: 0,
			wantFilePaths: []string{
				"package-lock.json", // TODO: remove when NPM is using the JSON matcher
				"package.json",
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tc := tt
			stdoutBuffer := &bytes.Buffer{}
			stderrBuffer := &bytes.Buffer{}

			ec := run(tc.args, stdoutBuffer, stderrBuffer)

			stdout := stdoutBuffer.String()
			bom := cyclonedx.BOM{}
			err := json.NewDecoder(strings.NewReader(stdout)).Decode(&bom)
			require.NoError(t, err)

			if ec != tc.wantExitCode {
				t.Errorf("cli exited with code %d, not %d", ec, tc.wantExitCode)
			}
			filepaths := gatherFilepath(bom)
			for _, expectedLocation := range tc.wantFilePaths {
				assert.Contains(t, filepaths, expectedLocation)
			}
		})
	}
}

func TestRun_WithCycloneDX15(t *testing.T) {
	t.Parallel()
	args := []string{
		"",
		"-r",
		"--experimental-only-packages",
		"--format=cyclonedx-1-5",
		"--consider-scan-path-as-root",
		"./fixtures/integration-test-locks",
	}
	stdoutBuffer := &bytes.Buffer{}
	stderrBuffer := &bytes.Buffer{}

	ec := run(args, stdoutBuffer, stderrBuffer)

	if ec != 0 {
		require.Failf(t, "The run did not finish successfully", "Error code = %v ; Error = %v", ec, stderrBuffer.String())
	}

	stdout := stdoutBuffer.String()
	bom := cyclonedx.BOM{}
	err := json.NewDecoder(strings.NewReader(stdout)).Decode(&bom)
	require.NoError(t, err)

	expectedBom := cyclonedx.BOM{
		JSONSchema:  "http://cyclonedx.org/schema/bom-1.5.schema.json",
		BOMFormat:   cyclonedx.BOMFormat,
		SpecVersion: cyclonedx.SpecVersion1_5,
		Version:     1,
		Components: &[]cyclonedx.Component{
			{
				BOMRef:     "pkg:maven/com.google.code.findbugs/jsr305@3.0.2",
				PackageURL: "pkg:maven/com.google.code.findbugs/jsr305@3.0.2",
				Type:       "library",
				Name:       "com.google.code.findbugs:jsr305",
				Version:    "3.0.2",
				Evidence: buildLocationEvidence(t, models.PackageLocations{
					Block: models.PackageLocation{
						Filename:    "/pom.xml",
						LineStart:   25,
						LineEnd:     28,
						ColumnStart: 5,
						ColumnEnd:   18,
					},
					Name: &models.PackageLocation{
						Filename:    "/pom.xml",
						LineStart:   27,
						LineEnd:     27,
						ColumnStart: 19,
						ColumnEnd:   25,
					},
					Version: &models.PackageLocation{
						Filename:    "/pom.xml",
						LineStart:   19,
						LineEnd:     19,
						ColumnStart: 18,
						ColumnEnd:   23,
					},
				}),
			},
			{
				BOMRef:     "pkg:nuget/Test.Core@6.0.5",
				PackageURL: "pkg:nuget/Test.Core@6.0.5",
				Type:       "library",
				Name:       "Test.Core",
				Version:    "6.0.5",
			},
		},
	}
	sbom_test.AssertBomEqual(t, expectedBom, bom, true)
}

func TestRun_WithEmptyCycloneDX15(t *testing.T) {
	t.Parallel()
	args := []string{
		"",
		"-r",
		"--experimental-only-packages",
		"--format=cyclonedx-1-5",
		"--consider-scan-path-as-root",
		"./fixtures/locks-empty",
	}
	stdoutBuffer := &bytes.Buffer{}
	stderrBuffer := &bytes.Buffer{}

	ec := run(args, stdoutBuffer, stderrBuffer)

	if ec != 0 {
		require.Failf(t, "The run did not finish successfully", "Error code = %v ; Error = %v", ec, stderrBuffer.String())
	}

	stdout := testutility.NormalizeStdStream(t, stdoutBuffer)
	stderr := testutility.NormalizeStdStream(t, stderrBuffer)
	testutility.NewSnapshot().MatchText(t, stdout)
	testutility.NewSnapshot().MatchText(t, stderr)
}

func TestRun_WithExplicitParsers(t *testing.T) {
	t.Parallel()
	args := []string{
		"",
		"-r",
		"--experimental-only-packages",
		"--format=cyclonedx-1-5",
		"--consider-scan-path-as-root",
		"--enable-parsers=pom.xml",
		"./fixtures/integration-test-locks",
	}
	stdoutBuffer := &bytes.Buffer{}
	stderrBuffer := &bytes.Buffer{}

	ec := run(args, stdoutBuffer, stderrBuffer)

	if ec != 0 {
		require.Failf(t, "The run did not finish successfully", "Error code = %v ; Error = %v", ec, stderrBuffer.String())
	}

	stdout := stdoutBuffer.String()
	bom := cyclonedx.BOM{}
	err := json.NewDecoder(strings.NewReader(stdout)).Decode(&bom)
	require.NoError(t, err)

	expectedBom := cyclonedx.BOM{
		JSONSchema:  "http://cyclonedx.org/schema/bom-1.5.schema.json",
		BOMFormat:   cyclonedx.BOMFormat,
		SpecVersion: cyclonedx.SpecVersion1_5,
		Version:     1,
		Components: &[]cyclonedx.Component{
			{
				BOMRef:     "pkg:maven/com.google.code.findbugs/jsr305@3.0.2",
				PackageURL: "pkg:maven/com.google.code.findbugs/jsr305@3.0.2",
				Type:       "library",
				Name:       "com.google.code.findbugs:jsr305",
				Version:    "3.0.2",
				Evidence: buildLocationEvidence(t, models.PackageLocations{
					Block: models.PackageLocation{
						Filename:    "/pom.xml",
						LineStart:   25,
						LineEnd:     28,
						ColumnStart: 5,
						ColumnEnd:   18,
					},
					Name: &models.PackageLocation{
						Filename:    "/pom.xml",
						LineStart:   27,
						LineEnd:     27,
						ColumnStart: 19,
						ColumnEnd:   25,
					},
					Version: &models.PackageLocation{
						Filename:    "/pom.xml",
						LineStart:   19,
						LineEnd:     19,
						ColumnStart: 18,
						ColumnEnd:   23,
					},
				}),
			},
		},
	}
	sbom_test.AssertBomEqual(t, expectedBom, bom, true)
}

func TestRun_WithEncodedLockfile(t *testing.T) {
	t.Parallel()
	testCases := []encodingTestCase{
		{encoding: "UTF-8"},
		{encoding: "UTF-16"},
		{encoding: "Windows-1252"},
	}

	expectedBom := cyclonedx.BOM{
		JSONSchema:  "http://cyclonedx.org/schema/bom-1.5.schema.json",
		BOMFormat:   cyclonedx.BOMFormat,
		SpecVersion: cyclonedx.SpecVersion1_5,
		Version:     1,
		Components: &[]cyclonedx.Component{
			{
				BOMRef:     "pkg:cargo/addr2line@0.15.2",
				PackageURL: "pkg:cargo/addr2line@0.15.2",
				Type:       "library",
				Name:       "addr2line",
				Version:    "0.15.2",
				Evidence:   sbom_test.BuildEmptyEvidence(),
			},
			{
				BOMRef:     "pkg:composer/sentry/sdk@2.0.4",
				PackageURL: "pkg:composer/sentry/sdk@2.0.4",
				Type:       "library",
				Name:       "sentry/sdk",
				Version:    "2.0.4",
				Evidence:   sbom_test.BuildEmptyEvidence(),
			},
			{
				BOMRef:     "pkg:conan/zlib@1.2.11",
				PackageURL: "pkg:conan/zlib@1.2.11",
				Type:       "library",
				Name:       "zlib",
				Version:    "1.2.11",
				Evidence:   sbom_test.BuildEmptyEvidence(),
			},
			{
				BOMRef:     "pkg:golang/stdlib@1.21.3",
				PackageURL: "pkg:golang/stdlib@1.21.3",
				Type:       "library",
				Name:       "stdlib",
				Version:    "1.21.3",
				Evidence:   sbom_test.BuildEmptyEvidence(),
			},
			{
				BOMRef:     "pkg:golang/github.com/BurntSushi/toml@1.0.0",
				PackageURL: "pkg:golang/github.com/BurntSushi/toml@1.0.0",
				Type:       "library",
				Name:       "github.com/BurntSushi/toml",
				Version:    "1.0.0",
				Evidence: buildLocationEvidence(t, models.PackageLocations{
					Block: models.PackageLocation{
						Filename:    "go.mod",
						LineStart:   5,
						LineEnd:     5,
						ColumnStart: 1,
						ColumnEnd:   42,
					},
					Version: &models.PackageLocation{
						Filename:    "go.mod",
						LineStart:   5,
						LineEnd:     5,
						ColumnStart: 37,
						ColumnEnd:   42,
					},
					Name: &models.PackageLocation{
						Filename:    "go.mod",
						LineStart:   5,
						LineEnd:     5,
						ColumnStart: 9,
						ColumnEnd:   35,
					},
				}),
			},
			{
				BOMRef:     "pkg:maven/org.springframework.security/spring-security-crypto@5.7.3",
				PackageURL: "pkg:maven/org.springframework.security/spring-security-crypto@5.7.3",
				Type:       "library",
				Name:       "org.springframework.security:spring-security-crypto",
				Version:    "5.7.3",
				Evidence: buildLocationEvidence(t, models.PackageLocations{
					Block: models.PackageLocation{
						Filename:    "build.gradle",
						LineStart:   10,
						LineEnd:     10,
						ColumnStart: 3,
						ColumnEnd:   77,
					},
					Name: &models.PackageLocation{
						Filename:    "build.gradle",
						LineStart:   10,
						LineEnd:     10,
						ColumnStart: 48,
						ColumnEnd:   70,
					},
					Version: &models.PackageLocation{
						Filename:    "build.gradle",
						LineStart:   10,
						LineEnd:     10,
						ColumnStart: 71,
						ColumnEnd:   76,
					},
				}),
			},
			{
				BOMRef:     "pkg:hex/plug@1.11.1",
				PackageURL: "pkg:hex/plug@1.11.1",
				Type:       "library",
				Name:       "plug",
				Version:    "1.11.1",
				Evidence:   sbom_test.BuildEmptyEvidence(),
			},
			{
				BOMRef:     "pkg:npm/wrappy@1.0.2",
				PackageURL: "pkg:npm/wrappy@1.0.2",
				Type:       "library",
				Name:       "wrappy",
				Version:    "1.0.2",
				Evidence: buildLocationEvidence(t, models.PackageLocations{
					Block: models.PackageLocation{
						Filename:    "npm/package.json",
						LineStart:   4,
						LineEnd:     4,
						ColumnStart: 5,
						ColumnEnd:   18,
					},
					Name: &models.PackageLocation{
						Filename:    "npm/package.json",
						LineStart:   4,
						LineEnd:     4,
						ColumnStart: 6,
						ColumnEnd:   12,
					},
					Version: &models.PackageLocation{
						Filename:    "npm/package.json",
						LineStart:   4,
						LineEnd:     4,
						ColumnStart: 16,
						ColumnEnd:   17,
					},
				}),
			},
			{
				BOMRef:     "pkg:nuget/Test.Core@6.0.5",
				PackageURL: "pkg:nuget/Test.Core@6.0.5",
				Type:       "library",
				Name:       "Test.Core",
				Version:    "6.0.5",
				Evidence:   sbom_test.BuildEmptyEvidence(),
			},
			{
				BOMRef:     "pkg:pypi/markupsafe@2.1.1",
				PackageURL: "pkg:pypi/markupsafe@2.1.1",
				Type:       "library",
				Name:       "markupsafe",
				Version:    "2.1.1",
				Evidence: buildLocationEvidence(t, models.PackageLocations{
					Block: models.PackageLocation{
						Filename:    "Pipfile",
						LineStart:   7,
						LineEnd:     7,
						ColumnStart: 1,
						ColumnEnd:   25,
					},
					Version: &models.PackageLocation{
						Filename:    "Pipfile",
						LineStart:   7,
						LineEnd:     7,
						ColumnStart: 15,
						ColumnEnd:   24,
					},
					Name: &models.PackageLocation{
						Filename:    "Pipfile",
						LineStart:   7,
						LineEnd:     7,
						ColumnStart: 1,
						ColumnEnd:   11,
					},
				}),
			},
			{
				BOMRef:     "pkg:npm/acorn@8.7.0",
				PackageURL: "pkg:npm/acorn@8.7.0",
				Type:       "library",
				Name:       "acorn",
				Version:    "8.7.0",
				Evidence: buildLocationEvidence(t, models.PackageLocations{
					Block: models.PackageLocation{
						Filename:    "pnpm/package.json",
						LineStart:   4,
						LineEnd:     4,
						ColumnStart: 5,
						ColumnEnd:   22,
					},
					Name: &models.PackageLocation{
						Filename:    "pnpm/package.json",
						LineStart:   4,
						LineEnd:     4,
						ColumnStart: 6,
						ColumnEnd:   11,
					},
					Version: &models.PackageLocation{
						Filename:    "pnpm/package.json",
						LineStart:   4,
						LineEnd:     4,
						ColumnStart: 15,
						ColumnEnd:   21,
					},
				}),
			},
			{
				BOMRef:     "pkg:pypi/numpy@1.23.3",
				PackageURL: "pkg:pypi/numpy@1.23.3",
				Type:       "library",
				Name:       "numpy",
				Version:    "1.23.3",
				Evidence: buildLocationEvidence(t, models.PackageLocations{
					Block: models.PackageLocation{
						Filename:    "pyproject.toml",
						LineStart:   10,
						LineEnd:     10,
						ColumnStart: 1,
						ColumnEnd:   19,
					},
					Version: &models.PackageLocation{
						Filename:    "pyproject.toml",
						LineStart:   10,
						LineEnd:     10,
						ColumnStart: 10,
						ColumnEnd:   18,
					},
					Name: &models.PackageLocation{
						Filename:    "pyproject.toml",
						LineStart:   10,
						LineEnd:     10,
						ColumnStart: 1,
						ColumnEnd:   6,
					},
				}),
			},
			{
				BOMRef:     "pkg:maven/com.google.code.findbugs/jsr305@3.0.2",
				PackageURL: "pkg:maven/com.google.code.findbugs/jsr305@3.0.2",
				Type:       "library",
				Name:       "com.google.code.findbugs:jsr305",
				Version:    "3.0.2",
				Evidence: buildLocationEvidence(t, models.PackageLocations{
					Block: models.PackageLocation{
						Filename:    "pom.xml",
						LineStart:   25,
						LineEnd:     28,
						ColumnStart: 5,
						ColumnEnd:   18,
					},
					Name: &models.PackageLocation{
						Filename:    "pom.xml",
						LineStart:   27,
						LineEnd:     27,
						ColumnStart: 19,
						ColumnEnd:   25,
					},
					Version: &models.PackageLocation{
						Filename:    "pom.xml",
						LineStart:   19,
						LineEnd:     19,
						ColumnStart: 18,
						ColumnEnd:   23,
					},
				}),
			},
			{
				BOMRef:     "pkg:pub/back_button_interceptor@6.0.1",
				PackageURL: "pkg:pub/back_button_interceptor@6.0.1",
				Type:       "library",
				Name:       "back_button_interceptor",
				Version:    "6.0.1",
				Evidence:   sbom_test.BuildEmptyEvidence(),
			},
			{
				BOMRef:     "pkg:cran/morning@0.1.0",
				PackageURL: "pkg:cran/morning@0.1.0",
				Type:       "library",
				Name:       "morning",
				Version:    "0.1.0",
				Evidence:   sbom_test.BuildEmptyEvidence(),
			},
			{
				BOMRef:     "pkg:pypi/django@2.2.24",
				PackageURL: "pkg:pypi/django@2.2.24",
				Type:       "library",
				Name:       "django",
				Version:    "2.2.24",
				Evidence: buildLocationEvidence(t, models.PackageLocations{
					Block: models.PackageLocation{
						Filename:    "requirements.txt",
						LineStart:   1,
						LineEnd:     1,
						ColumnStart: 1,
						ColumnEnd:   15,
					},
					Version: &models.PackageLocation{
						Filename:    "requirements.txt",
						LineStart:   1,
						LineEnd:     1,
						ColumnStart: 9,
						ColumnEnd:   15,
					},
					Name: &models.PackageLocation{
						Filename:    "requirements.txt",
						LineStart:   1,
						LineEnd:     1,
						ColumnStart: 1,
						ColumnEnd:   7,
					},
				}),
			},
			{
				BOMRef:     "pkg:npm/balanced-match@1.0.2",
				PackageURL: "pkg:npm/balanced-match@1.0.2",
				Type:       "library",
				Name:       "balanced-match",
				Version:    "1.0.2",
				Evidence: buildLocationEvidence(t, models.PackageLocations{
					Block: models.PackageLocation{
						Filename:    "yarn/package.json",
						LineStart:   4,
						LineEnd:     4,
						ColumnStart: 5,
						ColumnEnd:   31,
					},
					Name: &models.PackageLocation{
						Filename:    "yarn/package.json",
						LineStart:   4,
						LineEnd:     4,
						ColumnStart: 6,
						ColumnEnd:   20,
					},
					Version: &models.PackageLocation{
						Filename:    "yarn/package.json",
						LineStart:   4,
						LineEnd:     4,
						ColumnStart: 24,
						ColumnEnd:   30,
					},
				}),
			},
		},
	}

	for _, testCase := range testCases {
		tt := testCase
		t.Run(tt.encoding, func(t *testing.T) {
			t.Parallel()
			args := []string{
				"",
				"-r",
				"--experimental-only-packages",
				"--format=cyclonedx-1-5",
				"--paths-relative-to-scan-dir",
				"./fixtures/encoding-integration-test-locks/" + tt.encoding,
			}
			stdoutBuffer := &bytes.Buffer{}
			stderrBuffer := &bytes.Buffer{}

			ec := run(args, stdoutBuffer, stderrBuffer)

			if ec != 0 {
				require.Failf(t, "The run did not finish successfully", "Error code = %v ; Error = %v", ec, stderrBuffer.String())
			}

			stdout := stdoutBuffer.String()
			bom := cyclonedx.BOM{}
			err := json.NewDecoder(strings.NewReader(stdout)).Decode(&bom)
			require.NoError(t, err)

			sbom_test.AssertBomEqual(t, expectedBom, bom, true)
		})
	}
}

func buildLocationEvidence(t *testing.T, packageLocations models.PackageLocations) *cyclonedx.Evidence {
	t.Helper()
	jsonLocation := strings.Builder{}
	require.NoError(t, json.NewEncoder(&jsonLocation).Encode(packageLocations))

	return &cyclonedx.Evidence{
		Occurrences: &[]cyclonedx.EvidenceOccurrence{
			{
				Location: jsonLocation.String(),
			},
		},
	}
}

func gatherFilepath(bom cyclonedx.BOM) []string {
	locations := make([]string, 0)
	for _, component := range *bom.Components {
		for _, location := range *component.Evidence.Occurrences {
			jsonLocation := make(map[string]map[string]interface{})
			_ = json.NewDecoder(strings.NewReader(location.Location)).Decode(&jsonLocation)
			blockLocation := jsonLocation["block"]
			locations = append(locations, blockLocation["file_name"].(string))
		}
	}

	return locations
}

func TestRun_OCIImage(t *testing.T) {
	t.Parallel()

	testutility.SkipIfNotAcceptanceTesting(t, "Not consistent on MacOS/Windows")

	tests := []cliTestCase{
		{
			name: "Invalid path",
			args: []string{"", "--experimental-oci-image", "./fixtures/oci-image/no-file-here.tar"},
			exit: 127,
		},
		{
			name: "Alpine 3.10 image tar with 3.18 version file",
			args: []string{"", "--experimental-oci-image", "../../internal/image/fixtures/test-alpine.tar"},
			exit: 0,
		},
		{
			name: "scanning node_modules using npm with no packages",
			args: []string{"", "--experimental-oci-image", "../../internal/image/fixtures/test-node_modules-npm-empty.tar"},
			exit: 0,
		},
		{
			name: "scanning node_modules using npm with some packages",
			args: []string{"", "--experimental-oci-image", "../../internal/image/fixtures/test-node_modules-npm-full.tar"},
			exit: 0,
		},
		{
			name: "scanning node_modules using yarn with no packages",
			args: []string{"", "--experimental-oci-image", "../../internal/image/fixtures/test-node_modules-yarn-empty.tar"},
			exit: 0,
		},
		{
			name: "scanning node_modules using yarn with some packages",
			args: []string{"", "--experimental-oci-image", "../../internal/image/fixtures/test-node_modules-yarn-full.tar"},
			exit: 0,
		},
		{
			name: "scanning node_modules using pnpm with no packages",
			args: []string{"", "--experimental-oci-image", "../../internal/image/fixtures/test-node_modules-pnpm-empty.tar"},
			exit: 0,
		},
		{
			name: "scanning node_modules using pnpm with some packages",
			args: []string{"", "--experimental-oci-image", "../../internal/image/fixtures/test-node_modules-pnpm-full.tar"},
			exit: 0,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// point out that we need the images to be built and saved separately
			for _, arg := range tt.args {
				if strings.HasPrefix(arg, "../../internal/image/fixtures/") && strings.HasSuffix(arg, ".tar") {
					if _, err := os.Stat(arg); errors.Is(err, os.ErrNotExist) {
						t.Fatalf("%s does not exist - have you run scripts/build_test_images.sh?", arg)
					}
				}
			}

			testCli(t, tt)
		})
	}
}

// Tests all subcommands here.
func TestRun_SubCommands(t *testing.T) {
	t.Parallel()
	tests := []cliTestCase{
		// without subcommands
		{
			name: "with no subcommand",
			args: []string{"", "./fixtures/locks-many/composer.lock"},
			exit: 0,
		},
		// with scan subcommand
		{
			name: "with scan subcommand",
			args: []string{"", "scan", "./fixtures/locks-many/composer.lock"},
			exit: 0,
		},
		// scan with a flag
		{
			name: "scan with a flag",
			args: []string{"", "scan", "--recursive", "./fixtures/locks-one-with-nested"},
			exit: 0,
		},
		// TODO: add tests for other future subcommands
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			testCli(t, tt)
		})
	}
}

func TestRun_InsertDefaultCommand(t *testing.T) {
	t.Parallel()
	commands := []*cli.Command{
		{Name: "default"},
		{Name: "scan"},
	}
	defaultCommand := "default"

	tests := []struct {
		originalArgs []string
		wantArgs     []string
	}{
		// test when default command is specified
		{
			originalArgs: []string{"", "default", "file"},
			wantArgs:     []string{"", "default", "file"},
		},
		// test when command is not specified
		{
			originalArgs: []string{"", "file"},
			wantArgs:     []string{"", "default", "file"},
		},
		// test when command is also a filename
		{
			originalArgs: []string{"", "scan"}, // `scan` exists as a file on filesystem (`./cmd/osv-scanner/scan`)
			wantArgs:     []string{"", "scan"},
		},
		// test when command is not valid
		{
			originalArgs: []string{"", "invalid"},
			wantArgs:     []string{"", "default", "invalid"},
		},
		// test when command is a built-in option
		{
			originalArgs: []string{"", "--version"},
			wantArgs:     []string{"", "--version"},
		},
		{
			originalArgs: []string{"", "-h"},
			wantArgs:     []string{"", "-h"},
		},
		{
			originalArgs: []string{"", "help"},
			wantArgs:     []string{"", "help"},
		},
	}

	for _, tt := range tests {
		tt := tt
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		argsActual := insertDefaultCommand(tt.originalArgs, commands, defaultCommand, stdout, stderr)
		if !reflect.DeepEqual(argsActual, tt.wantArgs) {
			t.Errorf("Test Failed. Details:\n"+
				"Args (Got):  %s\n"+
				"Args (Want): %s\n", argsActual, tt.wantArgs)
		}
		testutility.NewSnapshot().MatchText(t, testutility.NormalizeStdStream(t, stdout))
		testutility.NewSnapshot().MatchText(t, testutility.NormalizeStdStream(t, stderr))
	}
}
