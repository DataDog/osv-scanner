package lockfile_test

import (
	"bytes"
	"errors"
	"github.com/stretchr/testify/assert"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/osv-scanner/pkg/lockfile"
)

func TestParseYarnLock_v1_FileDoesNotExist(t *testing.T) {
	t.Parallel()

	packages, err := lockfile.ParseYarnLock("fixtures/yarn/does-not-exist")

	expectErrIs(t, err, fs.ErrNotExist)
	expectPackagesWithoutLocations(t, packages, []lockfile.PackageDetails{})
}

func TestParseYarnLock_v1_NoPackages(t *testing.T) {
	t.Parallel()

	packages, err := lockfile.ParseYarnLock("fixtures/yarn/empty.v1.lock")

	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	expectPackagesWithoutLocations(t, packages, []lockfile.PackageDetails{})
}

func TestParseYarnLock_v1_OnePackage(t *testing.T) {
	t.Parallel()
	dir, err := os.Getwd()
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	path := filepath.FromSlash(filepath.Join(dir, "fixtures/yarn/one-package.v1.lock"))
	packages, err := lockfile.ParseYarnLock(path)
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	expectPackagesWithoutLocations(t, packages, []lockfile.PackageDetails{
		{
			Name:           "balanced-match",
			Version:        "1.0.2",
			TargetVersions: []string{"^1.0.0"},
			Ecosystem:      lockfile.YarnEcosystem,
			CompareAs:      lockfile.YarnEcosystem,
		},
	})
}

func TestParseYarnLock_v1_OnePackage_MatcherFailed(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	stderr := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}
	os.Stderr = w

	// Mock matcher to fail
	matcherError := errors.New("matcher failed")
	lockfile.YarnExtractor.Matcher = FailingMatcher{Error: matcherError}

	path := filepath.FromSlash(filepath.Join(dir, "fixtures/yarn/one-package.v1.lock"))
	packages, err := lockfile.ParseYarnLock(path)
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	// Capture stderr
	_ = w.Close()
	os.Stderr = stderr
	var buffer bytes.Buffer
	_, err = io.Copy(&buffer, r)
	if err != nil {
		t.Errorf("failed to copy stderr output: %v", err)
	}
	_ = r.Close()

	assert.Contains(t, buffer.String(), matcherError.Error())
	expectPackagesWithoutLocations(t, packages, []lockfile.PackageDetails{
		{
			Name:           "balanced-match",
			Version:        "1.0.2",
			TargetVersions: []string{"^1.0.0"},
			Ecosystem:      lockfile.YarnEcosystem,
			CompareAs:      lockfile.YarnEcosystem,
		},
	})

	// Reset matcher mock
	MockAllMatchers()
}

func TestParseYarnLock_v1_TwoPackages(t *testing.T) {
	t.Parallel()
	dir, err := os.Getwd()
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	path := filepath.FromSlash(filepath.Join(dir, "fixtures/yarn/two-packages.v1.lock"))
	packages, err := lockfile.ParseYarnLock(path)
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	expectPackagesWithoutLocations(t, packages, []lockfile.PackageDetails{
		{
			Name:           "concat-stream",
			Version:        "1.6.2",
			TargetVersions: []string{"^1.5.0"},
			Ecosystem:      lockfile.YarnEcosystem,
			CompareAs:      lockfile.YarnEcosystem,
		},
		{
			Name:           "concat-map",
			Version:        "0.0.1",
			TargetVersions: []string{"0.0.1"},
			Ecosystem:      lockfile.YarnEcosystem,
			CompareAs:      lockfile.YarnEcosystem,
		},
	})
}

func TestParseYarnLock_v1_WithQuotes(t *testing.T) {
	t.Parallel()
	dir, err := os.Getwd()
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	path := filepath.FromSlash(filepath.Join(dir, "fixtures/yarn/with-quotes.v1.lock"))
	packages, err := lockfile.ParseYarnLock(path)
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	expectPackagesWithoutLocations(t, packages, []lockfile.PackageDetails{
		{
			Name:           "concat-stream",
			Version:        "1.6.2",
			TargetVersions: []string{"^1.5.0"},
			Ecosystem:      lockfile.YarnEcosystem,
			CompareAs:      lockfile.YarnEcosystem,
		},
		{
			Name:           "concat-map",
			Version:        "0.0.1",
			TargetVersions: []string{"0.0.1"},
			Ecosystem:      lockfile.YarnEcosystem,
			CompareAs:      lockfile.YarnEcosystem,
		},
	})
}

func TestParseYarnLock_v1_MultipleVersions(t *testing.T) {
	t.Parallel()
	dir, err := os.Getwd()
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	path := filepath.FromSlash(filepath.Join(dir, "fixtures/yarn/multiple-versions.v1.lock"))
	packages, err := lockfile.ParseYarnLock(path)
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	expectPackagesWithoutLocations(t, packages, []lockfile.PackageDetails{
		{
			Name:           "define-properties",
			Version:        "1.1.3",
			TargetVersions: []string{"^1.1.3"},
			Ecosystem:      lockfile.YarnEcosystem,
			CompareAs:      lockfile.YarnEcosystem,
		},
		{
			Name:           "define-property",
			Version:        "0.2.5",
			TargetVersions: []string{"^0.2.5"},
			Ecosystem:      lockfile.YarnEcosystem,
			CompareAs:      lockfile.YarnEcosystem,
		},
		{
			Name:           "define-property",
			Version:        "1.0.0",
			TargetVersions: []string{"^1.0.0"},
			Ecosystem:      lockfile.YarnEcosystem,
			CompareAs:      lockfile.YarnEcosystem,
		},
		{
			Name:           "define-property",
			Version:        "2.0.2",
			TargetVersions: []string{"^2.0.2"},
			Ecosystem:      lockfile.YarnEcosystem,
			CompareAs:      lockfile.YarnEcosystem,
		},
	})
}

func TestParseYarnLock_v1_MultipleConstraints(t *testing.T) {
	t.Parallel()
	dir, err := os.Getwd()
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	path := filepath.FromSlash(filepath.Join(dir, "fixtures/yarn/multiple-constraints.v1.lock"))
	packages, err := lockfile.ParseYarnLock(path)
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	expectPackagesWithoutLocations(t, packages, []lockfile.PackageDetails{
		{
			Name:           "@babel/code-frame",
			Version:        "7.12.13",
			TargetVersions: []string{"^7.0.0", "^7.12.13"},
			Ecosystem:      lockfile.YarnEcosystem,
			CompareAs:      lockfile.YarnEcosystem,
		},
		{
			Name:           "domelementtype",
			Version:        "1.3.1",
			TargetVersions: []string{"1", "^1.3.1"},
			Ecosystem:      lockfile.YarnEcosystem,
			CompareAs:      lockfile.YarnEcosystem,
		},
	})
}

func TestParseYarnLock_v1_ScopedPackages(t *testing.T) {
	t.Parallel()
	dir, err := os.Getwd()
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	path := filepath.FromSlash(filepath.Join(dir, "fixtures/yarn/scoped-packages.v1.lock"))
	packages, err := lockfile.ParseYarnLock(path)
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	expectPackagesWithoutLocations(t, packages, []lockfile.PackageDetails{
		{
			Name:           "@babel/code-frame",
			Version:        "7.12.11",
			TargetVersions: []string{"7.12.11"},
			Ecosystem:      lockfile.YarnEcosystem,
			CompareAs:      lockfile.YarnEcosystem,
		},
		{
			Name:           "@babel/compat-data",
			Version:        "7.14.0",
			TargetVersions: []string{"^7.13.11"},
			Ecosystem:      lockfile.YarnEcosystem,
			CompareAs:      lockfile.YarnEcosystem,
		},
	})
}

func TestParseYarnLock_v1_WithPrerelease(t *testing.T) {
	t.Parallel()
	dir, err := os.Getwd()
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	path := filepath.FromSlash(filepath.Join(dir, "fixtures/yarn/with-prerelease.v1.lock"))
	packages, err := lockfile.ParseYarnLock(path)
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	expectPackagesWithoutLocations(t, packages, []lockfile.PackageDetails{
		{
			Name:           "css-tree",
			Version:        "1.0.0-alpha.37",
			TargetVersions: []string{"1.0.0-alpha.37"},
			Ecosystem:      lockfile.YarnEcosystem,
			CompareAs:      lockfile.YarnEcosystem,
		},
		{
			Name:           "gensync",
			Version:        "1.0.0-beta.2",
			TargetVersions: []string{"^1.0.0-beta.2"},
			Ecosystem:      lockfile.YarnEcosystem,
			CompareAs:      lockfile.YarnEcosystem,
		},
		{
			Name:           "node-fetch",
			Version:        "3.0.0-beta.9",
			TargetVersions: []string{"3.0.0-beta.9"},
			Ecosystem:      lockfile.YarnEcosystem,
			CompareAs:      lockfile.YarnEcosystem,
		},
		{
			Name:           "resolve",
			Version:        "1.20.0",
			TargetVersions: []string{"^1.1.7", "^1.10.0", "^1.12.0", "^1.14.2", "^1.20.0"},
			Ecosystem:      lockfile.YarnEcosystem,
			CompareAs:      lockfile.YarnEcosystem,
		},
		{
			Name:           "resolve",
			Version:        "2.0.0-next.3",
			TargetVersions: []string{"^2.0.0-next.3"},
			Ecosystem:      lockfile.YarnEcosystem,
			CompareAs:      lockfile.YarnEcosystem,
		},
	})
}

func TestParseYarnLock_v1_WithBuildString(t *testing.T) {
	t.Parallel()
	dir, err := os.Getwd()
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	path := filepath.FromSlash(filepath.Join(dir, "fixtures/yarn/with-build-string.v1.lock"))
	packages, err := lockfile.ParseYarnLock(path)
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	expectPackagesWithoutLocations(t, packages, []lockfile.PackageDetails{
		{
			Name:           "domino",
			Version:        "2.1.6+git",
			TargetVersions: []string{"https://github.com/angular/domino.git#f2435fe1f9f7c91ade0bd472c4723e5eacd7d19a"},
			Ecosystem:      lockfile.YarnEcosystem,
			CompareAs:      lockfile.YarnEcosystem,
		},
		{
			Name:           "tslib",
			Version:        "2.6.2",
			TargetVersions: []string{"^2.3.0"},
			Ecosystem:      lockfile.YarnEcosystem,
			CompareAs:      lockfile.YarnEcosystem,
		},
	})
}

func TestParseYarnLock_v1_Commits(t *testing.T) {
	t.Parallel()
	dir, err := os.Getwd()
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	path := filepath.FromSlash(filepath.Join(dir, "fixtures/yarn/commits.v1.lock"))
	packages, err := lockfile.ParseYarnLock(path)
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	expectPackagesWithoutLocations(t, packages, []lockfile.PackageDetails{
		{
			Name:           "mine1",
			Version:        "1.0.0-alpha.37",
			TargetVersions: []string{"git+ssh://git@github.com:G-Rath/npm-git-repo-2#0a2d2506c1"},
			Ecosystem:      lockfile.YarnEcosystem,
			CompareAs:      lockfile.YarnEcosystem,
			Commit:         "0a2d2506c1fe299691fc5db53a2097db3bd615bc",
		},
		{
			Name:           "mine2",
			Version:        "0.0.1",
			TargetVersions: []string{"G-Rath/npm-git-repo-2#0a2d2506c1"},
			Ecosystem:      lockfile.YarnEcosystem,
			CompareAs:      lockfile.YarnEcosystem,
			Commit:         "0a2d2506c1fe299691fc5db53a2097db3bd615bc",
		},
		{
			Name:           "mine3",
			Version:        "1.2.3",
			TargetVersions: []string{"G-Rath/npm-git-repo-1"},
			Ecosystem:      lockfile.YarnEcosystem,
			CompareAs:      lockfile.YarnEcosystem,
			Commit:         "094e581aaf927d010e4b61d706ba584551dac502",
		},
		{
			Name:    "mine4",
			Version: "0.0.2",
			TargetVersions: []string{
				"git+ssh://git@github.com:G-Rath/npm-git-repo-2#another-branch",
				"git+ssh://git@github.com:G-Rath/npm-git-repo-2#aa3bdfcb",
			},
			Ecosystem: lockfile.YarnEcosystem,
			CompareAs: lockfile.YarnEcosystem,
			Commit:    "aa3bdfcb1d845c79f14abb66f60d35b8a3ee5998",
		},
		{
			Name:           "mine4",
			Version:        "0.0.4",
			TargetVersions: []string{"G-Rath/npm-git-repo-2#another-branch"},
			Ecosystem:      lockfile.YarnEcosystem,
			CompareAs:      lockfile.YarnEcosystem,
			Commit:         "aa3bdfcb1d845c79f14abb66f60d35b8a3ee5998",
		},
		{
			Name:           "my-package",
			Version:        "1.8.3",
			TargetVersions: []string{"git+https://git@github.com/my-org/my-package.git#v1.8.3"},
			Ecosystem:      lockfile.YarnEcosystem,
			CompareAs:      lockfile.YarnEcosystem,
			Commit:         "b3bd3f1b3dad036e671251f5258beaae398f983a",
		},
		{
			Name:           "@bower_components/angular-animate",
			Version:        "1.4.14",
			TargetVersions: []string{"git://github.com/angular/bower-angular-animate.git#~1.4.0"},
			Ecosystem:      lockfile.YarnEcosystem,
			CompareAs:      lockfile.YarnEcosystem,
			Commit:         "e7f778fc054a086ba3326d898a00fa1bc78650a8",
		},
		{
			Name:           "@bower_components/alertify",
			Version:        "0.0.0",
			TargetVersions: []string{"fabien-d/alertify.js-shim#^0.3.10"},
			Ecosystem:      lockfile.YarnEcosystem,
			CompareAs:      lockfile.YarnEcosystem,
			Commit:         "e7b6c46d76604d297c389d830817b611c9a8f17c",
		},
		{
			Name:           "minimist",
			Version:        "0.0.8",
			TargetVersions: []string{"0.0.8", "ssh://github.com/substack/minimist.git#0.0.8"},
			Ecosystem:      lockfile.YarnEcosystem,
			CompareAs:      lockfile.YarnEcosystem,
			Commit:         "3754568bfd43a841d2d72d7fb54598635aea8fa4",
		},
		{
			Name:           "bats-assert",
			Version:        "2.0.0",
			TargetVersions: []string{"https://github.com/bats-core/bats-assert"},
			Ecosystem:      lockfile.YarnEcosystem,
			CompareAs:      lockfile.YarnEcosystem,
			Commit:         "4bdd58d3fbcdce3209033d44d884e87add1d8405",
		},
		{
			Name:           "bats-support",
			Version:        "0.3.0",
			TargetVersions: []string{"https://github.com/bats-core/bats-support"},
			Ecosystem:      lockfile.YarnEcosystem,
			CompareAs:      lockfile.YarnEcosystem,
			Commit:         "d140a65044b2d6810381935ae7f0c94c7023c8c3",
		},
		{
			Name:           "bats",
			Version:        "1.5.0",
			TargetVersions: []string{"https://github.com/bats-core/bats-core#master"},
			Ecosystem:      lockfile.YarnEcosystem,
			CompareAs:      lockfile.YarnEcosystem,
			Commit:         "172580d2ce19ee33780b5f1df817bbddced43789",
		},
		{
			Name:           "vue",
			Version:        "2.6.12",
			TargetVersions: []string{"https://github.com/vuejs/vue.git#v2.6.12"},
			Ecosystem:      lockfile.YarnEcosystem,
			CompareAs:      lockfile.YarnEcosystem,
			Commit:         "bb253db0b3e17124b6d1fe93fbf2db35470a1347",
		},
		{
			Name:           "kit",
			Version:        "1.0.0",
			TargetVersions: []string{"git+https://bitbucket.org/kettlelogic/kit.git"},
			Ecosystem:      lockfile.YarnEcosystem,
			CompareAs:      lockfile.YarnEcosystem,
			Commit:         "5b6830c0252eb73c6024d40a8ff5106d3023a2a6",
		},
		{
			Name:           "casadistance",
			Version:        "1.0.0",
			TargetVersions: []string{"git+ssh://git@bitbucket.org/casasoftag/casadistance.git"},
			Ecosystem:      lockfile.YarnEcosystem,
			CompareAs:      lockfile.YarnEcosystem,
			Commit:         "f0308391f0c50104182bfb2332a53e4e523a4603",
		},
		{
			Name:           "babel-preset-php",
			Version:        "1.1.1",
			TargetVersions: []string{"gitlab:kornelski/babel-preset-php#master"},
			Ecosystem:      lockfile.YarnEcosystem,
			CompareAs:      lockfile.YarnEcosystem,
			Commit:         "c5a7ba5e0ad98b8db1cb8ce105403dd4b768cced",
		},
		{
			Name:           "is-number",
			Version:        "2.0.0",
			TargetVersions: []string{"github:jonschlinkert/is-number#master"},
			Ecosystem:      lockfile.YarnEcosystem,
			CompareAs:      lockfile.YarnEcosystem,
			Commit:         "d5ac0584ee9ae7bd9288220a39780f155b9ad4c8",
		},
		{
			Name:           "is-number",
			Version:        "5.0.0",
			TargetVersions: []string{"https://dummy-token@github.com/jonschlinkert/is-number.git#master"},
			Ecosystem:      lockfile.YarnEcosystem,
			CompareAs:      lockfile.YarnEcosystem,
			Commit:         "af885e2e890b9ef0875edd2b117305119ee5bdc5",
		},
	})
}

func TestParseYarnLock_v1_Files(t *testing.T) {
	t.Parallel()
	dir, err := os.Getwd()
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	path := filepath.FromSlash(filepath.Join(dir, "fixtures/yarn/files.v1.lock"))
	packages, err := lockfile.ParseYarnLock(path)
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	expectPackagesWithoutLocations(t, packages, []lockfile.PackageDetails{
		{
			Name:           "etag",
			Version:        "1.8.1",
			TargetVersions: []string{"^1.0.0"},
			Ecosystem:      lockfile.YarnEcosystem,
			CompareAs:      lockfile.YarnEcosystem,
			Commit:         "",
		},
		{
			Name:           "filedep",
			Version:        "1.2.0",
			TargetVersions: []string{"../../correct/path/filedep"},
			Ecosystem:      lockfile.YarnEcosystem,
			CompareAs:      lockfile.YarnEcosystem,
			Commit:         "",
		},
		{
			Name:           "lodash",
			Version:        "1.3.1",
			TargetVersions: []string{"^1.3.1"},
			Ecosystem:      lockfile.YarnEcosystem,
			CompareAs:      lockfile.YarnEcosystem,
			Commit:         "",
		},
		{
			Name:           "other_package",
			Version:        "0.0.2",
			TargetVersions: []string{"./deps/other_package"},
			Ecosystem:      lockfile.YarnEcosystem,
			CompareAs:      lockfile.YarnEcosystem,
			Commit:         "",
		},
		{
			Name:           "sprintf-js",
			Version:        "0.0.0",
			TargetVersions: []string{"./mocks/sprintf-js", "~1.0.2"},
			Ecosystem:      lockfile.YarnEcosystem,
			CompareAs:      lockfile.YarnEcosystem,
			Commit:         "",
		},
		{
			Name:           "etag",
			Version:        "1.8.0",
			TargetVersions: []string{"./deps/etag"},
			Ecosystem:      lockfile.YarnEcosystem,
			CompareAs:      lockfile.YarnEcosystem,
			Commit:         "",
		},
	})
}

func TestParseYarnLock_v1_WithAliases(t *testing.T) {
	t.Parallel()
	dir, err := os.Getwd()
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	path := filepath.FromSlash(filepath.Join(dir, "fixtures/yarn/with-aliases.v1.lock"))
	packages, err := lockfile.ParseYarnLock(path)
	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	expectPackagesWithoutLocations(t, packages, []lockfile.PackageDetails{
		{
			Name:           "@babel/helper-validator-identifier",
			Version:        "7.22.20",
			TargetVersions: []string{"^7.0.0"},
			Ecosystem:      lockfile.YarnEcosystem,
			CompareAs:      lockfile.YarnEcosystem,
		},
		{
			Name:           "ansi-regex",
			Version:        "6.0.1",
			TargetVersions: []string{"^6.0.0"},
			Ecosystem:      lockfile.YarnEcosystem,
			CompareAs:      lockfile.YarnEcosystem,
		},
		{
			Name:           "ansi-regex",
			Version:        "5.0.1",
			TargetVersions: []string{"^5.0.0"},
			Ecosystem:      lockfile.YarnEcosystem,
			CompareAs:      lockfile.YarnEcosystem,
		},
	})
}
