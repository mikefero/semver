package semver

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const (
	numbers                string = "0123456789"
	alphas                        = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ-"
	alphanum                      = alphas + numbers
	kongVersionRegexFormat        = `^(?P<major>0|[1-9]\d*)\.` +
		`(?P<minor>0|[1-9]\d*)\.` +
		`(?P<patch>0|[1-9]\d*)` +
		`(?:\.(?P<revision>0|[1-9]\d*))?` +
		`(?:-(?P<prerelease>(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?` +
		`(?:\+(?P<buildmetadata>[0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`
	kongTolerantVersionRegexFormat = `^(?:v|\s*)?(?P<major>\d+)?` +
		`(?:\.)?(?P<minor>\d+)?` +
		`(?:\.)?(?P<patch>\d+)?` +
		`(?:\.(?P<revision>\d+))?` +
		`(?:-(?P<prerelease>(?:\d+|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:\d+|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?` +
		`(?:\+(?P<buildmetadata>[0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?\s*$`
)

var (
	// SpecVersion is the latest fully supported spec version of semver
	SpecVersion = Version{
		Major: 2,
		Minor: 0,
		Patch: 0,
	}

	KongVersionRegex         = regexp.MustCompile(kongVersionRegexFormat)
	KongTolerantVersionRegex = regexp.MustCompile(kongTolerantVersionRegexFormat)
)

// Version represents a semver compatible version
type Version struct {
	Major    uint64
	Minor    uint64
	Patch    uint64
	Revision int64
	Pre      []PRVersion
	Build    []string //No Precedence
}

// KongVersion to string
func (v Version) String() string {
	b := make([]byte, 0, 7)
	b = strconv.AppendUint(b, v.Major, 10)
	b = append(b, '.')
	b = strconv.AppendUint(b, v.Minor, 10)
	b = append(b, '.')
	b = strconv.AppendUint(b, v.Patch, 10)
	if v.Revision >= 0 {
		b = append(b, '.')
		b = strconv.AppendInt(b, v.Revision, 10)
	}

	if len(v.Pre) > 0 {
		b = append(b, '-')
		b = append(b, v.Pre[0].String()...)

		for _, pre := range v.Pre[1:] {
			b = append(b, '.')
			b = append(b, pre.String()...)
		}
	}

	if len(v.Build) > 0 {
		b = append(b, '+')
		b = append(b, v.Build[0]...)

		for _, build := range v.Build[1:] {
			b = append(b, '.')
			b = append(b, build...)
		}
	}

	return string(b)
}

// FinalizeVersion discards prerelease and build number and only returns
// major, minor and patch number.
func (v Version) FinalizeVersion() string {
	b := make([]byte, 0, 7)
	b = strconv.AppendUint(b, v.Major, 10)
	b = append(b, '.')
	b = strconv.AppendUint(b, v.Minor, 10)
	b = append(b, '.')
	b = strconv.AppendUint(b, v.Patch, 10)
	if v.Revision >= 0 {
		b = append(b, '.')
		b = strconv.AppendInt(b, v.Revision, 10)
	}
	return string(b)
}

// Equals checks if v is equal to o.
func (v Version) Equals(o Version) bool {
	return (v.Compare(o) == 0)
}

// EQ checks if v is equal to o.
func (v Version) EQ(o Version) bool {
	return (v.Compare(o) == 0)
}

// NE checks if v is not equal to o.
func (v Version) NE(o Version) bool {
	return (v.Compare(o) != 0)
}

// GT checks if v is greater than o.
func (v Version) GT(o Version) bool {
	return (v.Compare(o) == 1)
}

// GTE checks if v is greater than or equal to o.
func (v Version) GTE(o Version) bool {
	return (v.Compare(o) >= 0)
}

// GE checks if v is greater than or equal to o.
func (v Version) GE(o Version) bool {
	return (v.Compare(o) >= 0)
}

// LT checks if v is less than o.
func (v Version) LT(o Version) bool {
	return (v.Compare(o) == -1)
}

// LTE checks if v is less than or equal to o.
func (v Version) LTE(o Version) bool {
	return (v.Compare(o) <= 0)
}

// LE checks if v is less than or equal to o.
func (v Version) LE(o Version) bool {
	return (v.Compare(o) <= 0)
}

// Compare compares Versions v to o:
// -1 == v is less than o
// 0 == v is equal to o
// 1 == v is greater than o
func (v Version) Compare(o Version) int {
	if v.Major != o.Major {
		if v.Major > o.Major {
			return 1
		}
		return -1
	}
	if v.Minor != o.Minor {
		if v.Minor > o.Minor {
			return 1
		}
		return -1
	}
	if v.Patch != o.Patch {
		if v.Patch > o.Patch {
			return 1
		}
		return -1
	}

	// Handle revision comparison for 3-digit and 4-digit versions.
	// 1.2.2.3 < 1.2.3 is true
	// 1.2.3 < 1.2.3.4 is true
	if v.Revision != o.Revision && v.Revision != -1 && o.Revision != -1 {
		if v.Revision > o.Revision {
			return 1
		}
		return -1
	}

	// Quick comparison if a version has no prerelease versions
	if len(v.Pre) == 0 && len(o.Pre) == 0 {
		return 0
	} else if len(v.Pre) == 0 && len(o.Pre) > 0 {
		return 1
	} else if len(v.Pre) > 0 && len(o.Pre) == 0 {
		return -1
	}

	i := 0
	for ; i < len(v.Pre) && i < len(o.Pre); i++ {
		if comp := v.Pre[i].Compare(o.Pre[i]); comp == 0 {
			continue
		} else if comp == 1 {
			return 1
		} else {
			return -1
		}
	}

	// If all pr versions are the equal but one has further prversion, this one greater
	if i == len(v.Pre) && i == len(o.Pre) {
		return 0
	} else if i == len(v.Pre) && i < len(o.Pre) {
		return -1
	} else {
		return 1
	}

}

// IncrementPatch increments the patch version
func (v *Version) IncrementRevision() error {
	if v.Revision >= 0 {
		v.Revision++
	}
	return nil
}

// IncrementPatch increments the patch version
func (v *Version) IncrementPatch() error {
	v.Patch++
	if v.Revision >= 0 {
		v.Revision = 0
	}
	return nil
}

// IncrementMinor increments the minor version
func (v *Version) IncrementMinor() error {
	v.Minor++
	v.Patch = 0
	if v.Revision >= 0 {
		v.Revision = 0
	}
	return nil
}

// IncrementMajor increments the major version
func (v *Version) IncrementMajor() error {
	v.Major++
	v.Minor = 0
	v.Patch = 0
	if v.Revision >= 0 {
		v.Revision = 0
	}
	return nil
}

// Validate validates v and returns error in case
func (v Version) Validate() error {
	// Major, Minor, Patch already validated using uint64

	for _, pre := range v.Pre {
		if !pre.IsNum { //Numeric prerelease versions already uint64
			if len(pre.VersionStr) == 0 {
				return fmt.Errorf("Prerelease can not be empty %q", pre.VersionStr)
			}
			if !containsOnly(pre.VersionStr, alphanum) {
				return fmt.Errorf("Invalid character(s) found in prerelease %q", pre.VersionStr)
			}
		}
	}

	for _, build := range v.Build {
		if len(build) == 0 {
			return fmt.Errorf("Build meta data can not be empty %q", build)
		}
		if !containsOnly(build, alphanum) {
			return fmt.Errorf("Invalid character(s) found in build meta data %q", build)
		}
	}

	return nil
}

// New is an alias for Parse and returns a pointer, parses version string and returns a validated Version or error
func New(s string) (*Version, error) {
	v, err := Parse(s)
	vp := &v
	return vp, err
}

// Make is an alias for Parse, parses version string and returns a validated Version or error
func Make(s string) (Version, error) {
	return Parse(s)
}

// ParseTolerant allows for certain version specifications that do not strictly adhere to semver
// specs to be parsed by this library. It does so by normalizing versions before passing them to
// Parse(). It currently trims spaces, removes a "v" prefix, adds a 0 patch number to versions
// with only major and minor components specified, and removes leading 0s.
func ParseTolerant(s string) (Version, error) {
	if !KongTolerantVersionRegex.MatchString(s) {
		return Version{}, fmt.Errorf("Invalid tolerant version: '%s'", s)
	}

	// Split into major.minor.patch.revision-pr+build and remove leading zeros from
	// major, minor, patch, and revision
	parts := KongTolerantVersionRegex.FindStringSubmatch(s)
	majorStr := parts[KongTolerantVersionRegex.SubexpIndex("major")]
	minorStr := parts[KongTolerantVersionRegex.SubexpIndex("minor")]
	patchStr := parts[KongTolerantVersionRegex.SubexpIndex("patch")]
	revisionStr := parts[KongTolerantVersionRegex.SubexpIndex("revision")]
	prereleaseStr := parts[KongTolerantVersionRegex.SubexpIndex("prerelease")]
	buildStr := parts[KongTolerantVersionRegex.SubexpIndex("buildmetadata")]
	if len(majorStr) > 1 {
		majorStr = strings.TrimLeft(majorStr, "0")
		if len(majorStr) == 0 {
			majorStr = "0"
		}
	}
	if len(minorStr) > 1 {
		minorStr = strings.TrimLeft(minorStr, "0")
		if len(minorStr) == 0 {
			minorStr = "0"
		}
	}
	if len(patchStr) > 1 {
		patchStr = strings.TrimLeft(patchStr, "0")
		if len(patchStr) == 0 {
			patchStr = "0"
		}
	}
	if len(revisionStr) > 1 {
		revisionStr = strings.TrimLeft(revisionStr, "0")
		if len(revisionStr) == 0 {
			revisionStr = "0"
		}
	}

	// Fill up shortened versions.
	patchLen := len(patchStr)
	minorLen := len(minorStr)
	majorLen := len(majorStr)
	if patchLen == 0 || minorLen == 0 || majorLen == 0 {
		if len(prereleaseStr) > 0 || len(buildStr) > 0 {
			return Version{}, errors.New("Short version cannot contain PreRelease/Build meta data")
		}
		if len(patchStr) == 0 {
			patchStr = "0"
		}
		if len(minorStr) == 0 {
			minorStr = "0"
		}
		if len(majorStr) == 0 {
			majorStr = "0"
		}
	}

	// Generate version to properly parse
	s = fmt.Sprintf("%s.%s.%s", majorStr, minorStr, patchStr)
	if len(revisionStr) > 0 {
		s = fmt.Sprintf("%s.%s", s, revisionStr)
	}
	if len(prereleaseStr) > 0 {
		s = fmt.Sprintf("%s-%s", s, prereleaseStr)
	}
	if len(buildStr) > 0 {
		s = fmt.Sprintf("%s+%s", s, buildStr)
	}

	return Parse(s)
}

// Parse parses version string and returns a validated Version or error
func Parse(s string) (Version, error) {
	if len(s) == 0 {
		return Version{}, errors.New("Version string empty")
	}
	if !KongVersionRegex.MatchString(s) {
		return Version{}, fmt.Errorf("Invalid version: '%s'", s)
	}

	// Split into major.minor.patch.revision-pr+build
	parts := KongVersionRegex.FindStringSubmatch(s)
	major, err := strconv.ParseUint(parts[KongVersionRegex.SubexpIndex("major")], 10, 64)
	if err != nil {
		return Version{}, err
	}
	minor, err := strconv.ParseUint(parts[KongVersionRegex.SubexpIndex("minor")], 10, 64)
	if err != nil {
		return Version{}, err
	}
	patch, err := strconv.ParseUint(parts[KongVersionRegex.SubexpIndex("patch")], 10, 64)
	if err != nil {
		return Version{}, err
	}
	revisionStr := parts[KongVersionRegex.SubexpIndex("revision")]
	prereleaseStr := parts[KongVersionRegex.SubexpIndex("prerelease")]
	buildStr := parts[KongVersionRegex.SubexpIndex("buildmetadata")]

	v := Version{}
	v.Major = major
	v.Minor = minor
	v.Patch = patch
	v.Patch = patch
	v.Revision = -1
	if len(revisionStr) > 0 {
		revision, err := strconv.ParseInt(revisionStr, 10, 64)
		if err != nil {
			return Version{}, err
		}
		v.Revision = revision
	}

	var build, prerelease []string
	if len(buildStr) > 0 {
		build = strings.Split(buildStr, ".")
	}
	if len(prereleaseStr) > 0 {
		prerelease = strings.Split(prereleaseStr, ".")
	}

	// Prerelease
	for _, prstr := range prerelease {
		parsedPR, err := NewPRVersion(prstr)
		if err != nil {
			return Version{}, err
		}
		v.Pre = append(v.Pre, parsedPR)
	}

	// Build meta data
	for _, str := range build {
		if len(str) == 0 {
			return Version{}, errors.New("Build meta data is empty")
		}
		if !containsOnly(str, alphanum) {
			return Version{}, fmt.Errorf("Invalid character(s) found in build meta data %q", str)
		}
		v.Build = append(v.Build, str)
	}

	return v, nil
}

// MustParse is like Parse but panics if the version cannot be parsed.
func MustParse(s string) Version {
	v, err := Parse(s)
	if err != nil {
		panic(`semver: Parse(` + s + `): ` + err.Error())
	}
	return v
}

// PRVersion represents a PreRelease Version
type PRVersion struct {
	VersionStr string
	VersionNum uint64
	IsNum      bool
}

// NewPRVersion creates a new valid prerelease version
func NewPRVersion(s string) (PRVersion, error) {
	if len(s) == 0 {
		return PRVersion{}, errors.New("Prerelease is empty")
	}
	v := PRVersion{}
	if containsOnly(s, numbers) {
		if hasLeadingZeroes(s) {
			return PRVersion{}, fmt.Errorf("Numeric PreRelease version must not contain leading zeroes %q", s)
		}
		num, err := strconv.ParseUint(s, 10, 64)

		// Might never be hit, but just in case
		if err != nil {
			return PRVersion{}, err
		}
		v.VersionNum = num
		v.IsNum = true
	} else if containsOnly(s, alphanum) {
		v.VersionStr = s
		v.IsNum = false
	} else {
		return PRVersion{}, fmt.Errorf("Invalid character(s) found in prerelease %q", s)
	}
	return v, nil
}

// IsNumeric checks if prerelease-version is numeric
func (v PRVersion) IsNumeric() bool {
	return v.IsNum
}

// Compare compares two PreRelease Versions v and o:
// -1 == v is less than o
// 0 == v is equal to o
// 1 == v is greater than o
func (v PRVersion) Compare(o PRVersion) int {
	if v.IsNum && !o.IsNum {
		return -1
	} else if !v.IsNum && o.IsNum {
		return 1
	} else if v.IsNum && o.IsNum {
		if v.VersionNum == o.VersionNum {
			return 0
		} else if v.VersionNum > o.VersionNum {
			return 1
		} else {
			return -1
		}
	} else { // both are Alphas
		if v.VersionStr == o.VersionStr {
			return 0
		} else if v.VersionStr > o.VersionStr {
			return 1
		} else {
			return -1
		}
	}
}

// PreRelease version to string
func (v PRVersion) String() string {
	if v.IsNum {
		return strconv.FormatUint(v.VersionNum, 10)
	}
	return v.VersionStr
}

func containsOnly(s string, set string) bool {
	return strings.IndexFunc(s, func(r rune) bool {
		return !strings.ContainsRune(set, r)
	}) == -1
}

func hasLeadingZeroes(s string) bool {
	return len(s) > 1 && s[0] == '0'
}

// NewBuildVersion creates a new valid build version
func NewBuildVersion(s string) (string, error) {
	if len(s) == 0 {
		return "", errors.New("Buildversion is empty")
	}
	if !containsOnly(s, alphanum) {
		return "", fmt.Errorf("Invalid character(s) found in build meta data %q", s)
	}
	return s, nil
}

// FinalizeVersion returns the major, minor and patch number only and discards
// prerelease and build number.
func FinalizeVersion(s string) (string, error) {
	v, err := Parse(s)
	if err != nil {
		return "", err
	}
	v.Pre = nil
	v.Build = nil

	finalVer := v.String()
	return finalVer, nil
}
