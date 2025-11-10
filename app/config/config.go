package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// Config はアプリケーションの設定を管理する
type Config struct {
	TfspecDir string
	EnvDirs   []string
	Verbose   bool
	NoFail    bool
}

// ConfigService は設定関連の処理を担当する
type ConfigService struct{}

func NewConfigService() *ConfigService {
	return &ConfigService{}
}

// LoadConfig は設定を読み込んで検証する
func (s *ConfigService) LoadConfig(envDirs []string, verbose, noFail bool) (*Config, error) {
	tfspecDir, err := s.setupTfspecDir()
	if err != nil {
		return nil, err
	}

	resolvedEnvDirs, err := s.resolveEnvDirs(envDirs)
	if err != nil {
		return nil, err
	}

	return &Config{
		TfspecDir: tfspecDir,
		EnvDirs:   resolvedEnvDirs,
		Verbose:   verbose,
		NoFail:    noFail,
	}, nil
}

// setupTfspecDir は.tfspecディレクトリの存在を確認し、パスを返す
// ディレクトリが存在しない場合は空文字を返す（ignoreルールなしで動作）
func (s *ConfigService) setupTfspecDir() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("現在のディレクトリを取得できませんでした: %w", err)
	}

	tfspecDir := filepath.Join(cwd, ".tfspec")
	if _, err := os.Stat(tfspecDir); os.IsNotExist(err) {
		// .tfspecディレクトリが存在しない場合は空文字を返す（ignoreルールなし）
		return "", nil
	}

	return tfspecDir, nil
}

// resolveEnvDirs は環境ディレクトリを解決する
func (s *ConfigService) resolveEnvDirs(envDirs []string) ([]string, error) {
	if len(envDirs) == 0 {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("現在のディレクトリを取得できませんでした: %w", err)
		}

		envDirs, err = s.detectEnvDirs(cwd)
		if err != nil {
			return nil, fmt.Errorf("環境ディレクトリの自動検出に失敗しました: %w", err)
		}
	}

	if len(envDirs) == 0 {
		return nil, fmt.Errorf("環境ディレクトリが見つかりませんでした\n" +
			"ヒント: .tf または .hcl ファイルを含むディレクトリを作成するか、コマンドライン引数で環境ディレクトリを指定してください")
	}

	fmt.Printf("対象環境: %v\n", envDirs)
	return envDirs, nil
}

// detectEnvDirs は環境ディレクトリを自動検出する
func (s *ConfigService) detectEnvDirs(baseDir string) ([]string, error) {
	var envDirs []string

	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() || entry.Name() == ".tfspec" {
			continue
		}

		envPath := filepath.Join(baseDir, entry.Name())
		hasTerraformFiles, err := s.hasTerraformFiles(envPath)
		if err != nil {
			continue // ディレクトリの読み込みエラーは無視して次へ
		}

		if hasTerraformFiles {
			envDirs = append(envDirs, envPath)
		}
	}

	return envDirs, nil
}

// hasTerraformFiles は指定ディレクトリに .tf または .hcl ファイルが存在するかチェックする
func (s *ConfigService) hasTerraformFiles(dir string) (bool, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			name := entry.Name()
			if filepath.Ext(name) == ".tf" || filepath.Ext(name) == ".hcl" {
				return true, nil
			}
		}
	}

	return false, nil
}