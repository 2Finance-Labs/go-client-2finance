.PHONY: help test sdk-structure-test go-test go-fmt dart-test dart-blockchain-e2e javascript-test typescript-test python-test php-test java-test cpp-test examples-test

help:
	@echo "Targets:"
	@echo "  test      Run all local non-live SDK checks"
	@echo "  sdk-structure-test Validate cross-language SDK structure"
	@echo "  go-test   Run Go SDK tests"
	@echo "  go-fmt    Format Go SDK files"
	@echo "  dart-test Run Dart SDK unit/contract tests"
	@echo "  javascript-test Run JavaScript SDK tests"
	@echo "  typescript-test Run TypeScript SDK typecheck"
	@echo "  python-test Run Python SDK tests"
	@echo "  php-test Run PHP SDK smoke test"
	@echo "  java-test Compile and run Java SDK smoke test"
	@echo "  cpp-test Compile and run C++ SDK smoke test"
	@echo "  examples-test Validate quickstart examples where tooling is available"
	@echo "  dart-blockchain-e2e Run Dart blockchain live E2E tests"

test: sdk-structure-test go-test dart-test javascript-test typescript-test python-test php-test java-test cpp-test examples-test

sdk-structure-test:
	node tools/validate-sdk-structure.mjs

go-test:
	$(MAKE) -C go test-fast

go-fmt:
	$(MAKE) -C go fmt

dart-test:
	cd dart/sdk && dart pub get
	cd dart/sdk && dart analyze
	cd dart/sdk && dart test
	cd dart/auth && dart test
	cd dart/blockchain && dart test test/two_finance_blockchain_client_unit_test.dart test/http_transport_test.dart test/spec_harness/spec_files_test.dart test/blockchain/utils test/blockchain/transaction/transaction_test.dart test/blockchain/log/log_test.dart test/blockchain/types/types_test.dart test/blockchain/keys/keys_test.dart test/blockchain/contract/lifecycle/lifecycle_test.dart test/blockchain/contract/walletV2/wallet_unit_test.dart test/wallet_manager/wallet_manager_test.dart

dart-blockchain-e2e:
	cd dart/blockchain && dart test

javascript-test:
	cd javascript && npm test

typescript-test:
	@if command -v npm >/dev/null 2>&1 && [ -d typescript/node_modules ]; then \
		cd typescript && npm test; \
	elif command -v tsc >/dev/null 2>&1; then \
		cd typescript && tsc --noEmit; \
	else \
		echo "typescript dependencies not installed; skipping TypeScript typecheck"; \
	fi

python-test:
	cd python && PYTHONPATH=. python3 -m unittest discover -s tests -v
	cd python && PYTHONPATH=. python3 -m py_compile examples/quickstart.py

php-test:
	@if command -v php >/dev/null 2>&1; then \
		cd php && php -l examples/quickstart.php && php tests/SdkClientTest.php; \
	else \
		echo "php not installed; skipping PHP smoke test"; \
	fi

java-test:
	@if command -v javac >/dev/null 2>&1; then \
		$(MAKE) -C java test; \
	else \
		echo "javac not installed; skipping Java compile smoke test"; \
	fi

cpp-test:
	@if command -v cmake >/dev/null 2>&1; then \
		cd cpp && cmake -S . -B build && cmake --build build && ctest --test-dir build --output-on-failure; \
	else \
		cd cpp && mkdir -p build && c++ -std=c++17 -Iinclude tests/sdk_client_test.cpp -o build/sdk_client_test && ./build/sdk_client_test; \
	fi

examples-test:
	node --check javascript/examples/quickstart.js
	node --check javascript/examples/request-options.js
	node --check javascript/examples/auth-client-credentials.js
	node --check javascript/examples/error-handling.js
	cd python && PYTHONPATH=. python3 -m py_compile examples/quickstart.py examples/request_options.py examples/auth_client_credentials.py examples/error_handling.py
	@if command -v php >/dev/null 2>&1; then cd php && php -l examples/quickstart.php && php -l examples/request_options.php && php -l examples/auth_client_credentials.php && php -l examples/error_handling.php; else echo "php not installed; skipping PHP example syntax check"; fi
	@if command -v c++ >/dev/null 2>&1; then c++ -std=c++17 -Icpp/include cpp/examples/quickstart.cpp -o /tmp/twofinance_cpp_quickstart && c++ -std=c++17 -Icpp/include cpp/examples/request_options.cpp -o /tmp/twofinance_cpp_request_options && c++ -std=c++17 -Icpp/include cpp/examples/auth_token_source.cpp -o /tmp/twofinance_cpp_auth_token_source && c++ -std=c++17 -Icpp/include cpp/examples/error_handling.cpp -o /tmp/twofinance_cpp_error_handling; else echo "c++ not installed; skipping C++ example compile"; fi
