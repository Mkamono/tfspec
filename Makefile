# Makefile for tfspec

# バイナリ名
BINARY_NAME := tfspec

# テストケースディレクトリ
TEST_DIRS := $(wildcard test/*/.)
TEST_CASES := $(notdir $(patsubst %/.,%,$(TEST_DIRS)))

# デフォルトターゲット
.PHONY: help
help: ## ヘルプメッセージを表示
	@echo "利用可能なコマンド:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.PHONY: build
build: ## tfspecバイナリをビルド
	@echo "Building $(BINARY_NAME)..."
	@go build -o $(BINARY_NAME)
	@echo "✅ Build completed: $(BINARY_NAME)"

.PHONY: test-all
test-all: build ## 全テストケースでreport.mdを生成（no-fail使用、--trim-cell付き）
	@echo "🚀 Running tests on all test cases with --no-fail and --trim-cell..."
	@echo "📁 Test cases found: $(words $(TEST_CASES)) cases"
	@echo "Test cases: $(TEST_CASES)"
	@echo ""
	@failed=0; \
	for testcase in $(TEST_CASES); do \
		echo "🔍 Testing: $$testcase"; \
		if [ -d "test/$$testcase" ]; then \
			cd test/$$testcase && \
			if ../../$(BINARY_NAME) check --no-fail --trim-cell -o; then \
				echo "✅ $$testcase: report.md generated successfully"; \
			else \
				echo "❌ $$testcase: failed to generate report.md"; \
				failed=$$((failed + 1)); \
			fi && \
			cd ../..; \
		else \
			echo "❌ $$testcase: directory not found"; \
			failed=$$((failed + 1)); \
		fi; \
		echo ""; \
	done; \
	echo "📊 Test Summary:"; \
	echo "   Total test cases: $(words $(TEST_CASES))"; \
	echo "   Failed cases: $$failed"; \
	if [ $$failed -eq 0 ]; then \
		echo "🎉 All tests passed!"; \
	else \
		echo "⚠️  $$failed test case(s) failed"; \
		exit 1; \
	fi

.PHONY: test-case
test-case: build ## 特定のテストケースでreport.mdを生成（例: make test-case CASE=basic_attribute_diff、--trim-cell付き）
ifndef CASE
	@echo "❌ エラー: CASE変数を指定してください"
	@echo "例: make test-case CASE=basic_attribute_diff"
	@echo "利用可能なケース: $(TEST_CASES)"
	@exit 1
endif
	@echo "🔍 Testing specific case: $(CASE)"
	@if [ -d "test/$(CASE)" ]; then \
		cd test/$(CASE) && \
		echo "Generating report for $(CASE)..." && \
		../../$(BINARY_NAME) check --no-fail --trim-cell -o && \
		echo "✅ $(CASE): report.md generated at test/$(CASE)/.tfspec/report.md" && \
		cd ../..; \
	else \
		echo "❌ Test case '$(CASE)' not found"; \
		echo "利用可能なケース: $(TEST_CASES)"; \
		exit 1; \
	fi

.PHONY: clean-reports
clean-reports: ## 全テストケースのreport.mdを削除
	@echo "🧹 Cleaning all report.md files..."
	@find test -name "report.md" -type f -delete
	@find test -path "*/.tfspec/report.md" -type f -delete 2>/dev/null || true
	@echo "✅ All report files cleaned"

.PHONY: show-reports
show-reports: ## 生成されたreport.mdファイルの一覧を表示
	@echo "📄 Generated report files:"
	@find test -name "report.md" -type f | sort

.PHONY: list-cases
list-cases: ## 利用可能なテストケース一覧を表示
	@echo "📋 Available test cases:"
	@for case in $(TEST_CASES); do echo "   - $$case"; done

.PHONY: validate-reports
validate-reports: ## 生成されたreport.mdファイルの内容を簡易チェック
	@echo "🔍 Validating generated reports..."
	@failed=0; \
	for testcase in $(TEST_CASES); do \
		report_file="test/$$testcase/.tfspec/report.md"; \
		if [ -f "$$report_file" ]; then \
			if grep -q "# Tfspec Check Results" "$$report_file"; then \
				echo "✅ $$testcase: valid report format"; \
			else \
				echo "❌ $$testcase: invalid report format"; \
				failed=$$((failed + 1)); \
			fi; \
		else \
			echo "❌ $$testcase: report.md not found"; \
			failed=$$((failed + 1)); \
		fi; \
	done; \
	if [ $$failed -eq 0 ]; then \
		echo "🎉 All reports are valid!"; \
	else \
		echo "⚠️  $$failed report(s) are invalid or missing"; \
		exit 1; \
	fi

.PHONY: clean
clean: clean-reports ## バイナリとレポートをクリーンアップ
	@echo "🧹 Cleaning binary and reports..."
	@rm -f $(BINARY_NAME)
	@echo "✅ Cleanup completed"

# 開発者向けコマンド
.PHONY: dev-test
dev-test: clean build test-all validate-reports ## 開発者向け: クリーン→ビルド→テスト→検証の完全サイクル
	@echo "🎉 Development test cycle completed successfully!"