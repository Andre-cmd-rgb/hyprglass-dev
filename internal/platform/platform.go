package platform

import (
	"bufio"
	"os"
	"runtime"
	"slices"
	"strings"
)

type Info struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	PrettyName string   `json:"prettyName"`
	IDLike     []string `json:"idLike"`
	ArchLike   bool     `json:"archLike"`
	CachyOS    bool     `json:"cachyOS"`
	GOOS       string   `json:"goos"`
}

func Read() Info {
	b, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return Info{GOOS: runtime.GOOS}
	}
	info := ParseOSRelease(string(b))
	info.GOOS = runtime.GOOS
	return info
}

func ParseOSRelease(data string) Info {
	vals := map[string]string{}
	s := bufio.NewScanner(strings.NewReader(data))
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" || strings.HasPrefix(line, "#") || !strings.Contains(line, "=") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		val = strings.Trim(val, "\"")
		vals[key] = val
	}
	id := strings.ToLower(vals["ID"])
	idLike := fieldsLower(vals["ID_LIKE"])
	info := Info{
		ID:         id,
		Name:       vals["NAME"],
		PrettyName: vals["PRETTY_NAME"],
		IDLike:     idLike,
		CachyOS:    id == "cachyos",
	}
	info.ArchLike = id == "arch" || id == "cachyos" || slices.Contains(idLike, "arch") || slices.Contains(idLike, "archlinux")
	return info
}

func (i Info) DisplayName() string {
	if i.PrettyName != "" {
		return i.PrettyName
	}
	if i.Name != "" {
		return i.Name
	}
	if i.ID != "" {
		return i.ID
	}
	return i.GOOS
}

func (i Info) PackageFile() string {
	if i.CachyOS {
		return "packages/cachyos-core.txt"
	}
	return "packages/arch-core.txt"
}

func fieldsLower(v string) []string {
	var out []string
	for _, f := range strings.Fields(v) {
		out = append(out, strings.ToLower(f))
	}
	return out
}
