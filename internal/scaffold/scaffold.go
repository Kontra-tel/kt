package scaffold

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
)

type Context struct {
	App          string
	Template     string
	Author       string
	ServiceUser  string
	ServiceGroup string
}

type Scaffolder struct{ FS embed.FS }

type TemplateInfo struct {
	Name string
	Desc string
}

type templateAlias struct {
	Target string
	Desc   string
}

var templateAliases = map[string]templateAlias{
	"service": {
		Target: "app",
		Desc:   "Single-service application package",
	},
}

func (s Scaffolder) Templates() ([]string, error) {
	infos, err := s.TemplatesWithDesc()
	if err != nil {
		return nil, err
	}
	names := make([]string, len(infos))
	for i, t := range infos {
		names[i] = t.Name
	}
	return names, nil
}

func (s Scaffolder) TemplatesWithDesc() ([]TemplateInfo, error) {
	entries, err := fs.ReadDir(s.FS, "templates/projects")
	if err != nil {
		return nil, err
	}
	var out []TemplateInfo
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		info := TemplateInfo{Name: e.Name()}
		if data, err := s.FS.ReadFile("templates/projects/" + e.Name() + "/template.yaml"); err == nil {
			for _, line := range strings.Split(string(data), "\n") {
				if k, v, ok := strings.Cut(strings.TrimSpace(line), ":"); ok && strings.TrimSpace(k) == "description" {
					info.Desc = strings.TrimSpace(v)
					break
				}
			}
		}
		if info.Name == "app" && info.Desc != "" {
			info.Desc += " (legacy name; use service)"
		}
		out = append(out, info)
	}
	for name, alias := range templateAliases {
		out = append(out, TemplateInfo{Name: name, Desc: alias.Desc})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

func (s Scaffolder) InstallTools(dest string, force bool) error {
	pairs := []struct{ src, dst string }{{"common/mk", ".kt/mk"}, {"common/scripts", ".kt/scripts"}}
	for _, p := range pairs {
		if err := copyTree(s.FS, p.src, filepath.Join(dest, p.dst), nil, force); err != nil {
			return err
		}
	}
	return nil
}

func (s Scaffolder) Init(dest string, ctx Context, force bool) error {
	if ctx.App == "" {
		return fmt.Errorf("app name is required")
	}
	if ctx.Template == "" {
		return fmt.Errorf("template is required")
	}
	if ctx.Author == "" {
		ctx.Author = gitAuthor()
	}
	if ctx.ServiceUser == "" {
		ctx.ServiceUser = ctx.App
	}
	if ctx.ServiceGroup == "" {
		ctx.ServiceGroup = ctx.ServiceUser
	}
	if err := os.MkdirAll(dest, 0755); err != nil {
		return err
	}
	if err := s.InstallTools(dest, force); err != nil {
		return err
	}
	base := "templates/projects/" + resolveTemplate(ctx.Template)
	if _, err := fs.Stat(s.FS, base); err != nil {
		return fmt.Errorf("unknown template %q", ctx.Template)
	}
	if err := copyTree(s.FS, base, dest, &ctx, force); err != nil {
		return err
	}
	vf := filepath.Join(dest, "version.txt")
	if _, err := os.Stat(vf); os.IsNotExist(err) {
		if err := os.WriteFile(vf, []byte("0.1.0\n"), 0644); err != nil {
			return err
		}
	}
	return chmodScripts(dest)
}

func copyTree(efs embed.FS, srcRoot, dstRoot string, ctx *Context, force bool) error {
	return fs.WalkDir(efs, srcRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == srcRoot {
			return nil
		}
		rel, _ := filepath.Rel(srcRoot, path)
		if strings.HasSuffix(rel, "template.yaml") {
			return nil
		}
		dstRel := rel
		if ctx != nil {
			dstRel = strings.ReplaceAll(dstRel, "app.service", ctx.App+".service")
			dstRel = strings.ReplaceAll(dstRel, "service.service", ctx.App+"-service.service")
			dstRel = strings.ReplaceAll(dstRel, "backend.service", ctx.App+"-backend.service")
			dstRel = strings.ReplaceAll(dstRel, "frontend.service", ctx.App+"-frontend.service")
			dstRel = strings.ReplaceAll(dstRel, "cmd/app", "cmd/"+ctx.App)
			dstRel = strings.ReplaceAll(dstRel, "deploy/run/app", "deploy/run/"+ctx.App)
			dstRel = strings.ReplaceAll(dstRel, "deploy/run/service", "deploy/run/"+ctx.App+"-service")
			dstRel = strings.ReplaceAll(dstRel, "deploy/run/backend", "deploy/run/"+ctx.App+"-backend")
			dstRel = strings.ReplaceAll(dstRel, "deploy/run/frontend", "deploy/run/"+ctx.App+"-frontend")
			// deploy/bin/app → deploy/bin/<app>; also handles app-backend and app-frontend suffixes
			dstRel = strings.ReplaceAll(dstRel, "deploy/bin/app", "deploy/bin/"+ctx.App)
		}
		if strings.HasSuffix(dstRel, ".tmpl") {
			dstRel = strings.TrimSuffix(dstRel, ".tmpl")
		}
		dst := filepath.Join(dstRoot, dstRel)
		if d.IsDir() {
			return os.MkdirAll(dst, 0755)
		}
		data, err := efs.ReadFile(path)
		if err != nil {
			return err
		}
		if ctx != nil && strings.HasSuffix(path, ".tmpl") {
			t, err := template.New(filepath.Base(path)).Parse(string(data))
			if err != nil {
				return err
			}
			var b strings.Builder
			if err := t.Execute(&b, ctx); err != nil {
				return err
			}
			data = []byte(b.String())
		}
		if !force {
			if _, err := os.Stat(dst); err == nil {
				return nil
			}
		}
		if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
			return err
		}
		mode := fs.FileMode(0644)
		if strings.HasSuffix(dst, ".sh") || strings.Contains(dst, "/deploy/bin/") || strings.Contains(dst, "/deploy/run/") {
			mode = 0755
		}
		return os.WriteFile(dst, data, mode)
	})
}

func gitAuthor() string {
	name, _ := exec.Command("git", "config", "user.name").Output()
	email, _ := exec.Command("git", "config", "user.email").Output()
	n := strings.TrimSpace(string(name))
	e := strings.TrimSpace(string(email))
	if n != "" && e != "" {
		return n + " <" + e + ">"
	}
	if n != "" {
		return n
	}
	return ""
}

func resolveTemplate(name string) string {
	if alias, ok := templateAliases[name]; ok {
		return alias.Target
	}
	return name
}

func chmodScripts(root string) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		if strings.HasSuffix(path, ".sh") || strings.Contains(path, "/deploy/bin/") || strings.Contains(path, "/deploy/run/") {
			return os.Chmod(path, 0755)
		}
		return nil
	})
}
