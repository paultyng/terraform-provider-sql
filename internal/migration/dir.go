package migration

import (
	"bufio"
	"io/ioutil"
	"path/filepath"
	"strings"
)

const SHMigSplit = "-- ==== DOWN ===="

type Options struct {
	StripLineComments bool
	SingleFileSplit   string
}

var defaultOptions = &Options{
	StripLineComments: true,
}

func ReadDir(dir string, opts *Options) ([]Migration, error) {
	if opts == nil {
		opts = defaultOptions
	}

	// readdir returns the list already sorted by
	// file name so no need to sort
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var migrations []Migration

	for _, file := range files {
		if file.IsDir() {
			// ignore child directories
			continue
		}

		fileName := file.Name()

		ext := filepath.Ext(fileName)
		if strings.ToLower(ext) != ".sql" {
			// only process .sql files
			continue
		}

		fileNameNoExt := strings.TrimSuffix(fileName, ext)

		raw, err := ioutil.ReadFile(filepath.Join(dir, fileName))
		if err != nil {
			return nil, err
		}

		switch {
		case opts.SingleFileSplit != "":
			parts := strings.SplitN(string(raw), opts.SingleFileSplit, 2)
			m := Migration{
				ID: fileNameNoExt,
				Up: cleanSQL(parts[0], opts.StripLineComments),
			}
			if len(parts) == 2 {
				m.Down = cleanSQL(parts[1], opts.StripLineComments)
			}
			migrations = append(migrations, m)
		default:
			directionExt := filepath.Ext(fileNameNoExt)
			id := strings.TrimSuffix(fileNameNoExt, directionExt)

			sql := cleanSQL(string(raw), opts.StripLineComments)

			var m *Migration
			for i := range migrations {
				existingMigration := &migrations[i]
				if existingMigration.ID != id {
					continue
				}

				m = existingMigration
			}

			if m == nil {
				migrations = append(migrations, Migration{
					ID: id,
				})
				m = &migrations[len(migrations)-1]
			}

			switch strings.ToLower(directionExt) {
			case ".up":
				m.Up = sql
			case ".down":
				m.Down = sql
			}
		}
	}

	return migrations, nil
}

func cleanSQL(sql string, stripLineComments bool) string {
	if stripLineComments {
		scanner := bufio.NewScanner(strings.NewReader(sql))
		lines := []string{}
		scanner.Split(bufio.ScanLines)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "--") {
				continue
			}
			lines = append(lines, line)
		}
		sql = strings.Join(lines, "\n")
	}
	sql = strings.TrimSpace(sql)
	return sql
}
