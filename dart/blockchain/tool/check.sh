#!/usr/bin/env bash
set -euo pipefail

export HOME="${CHECK_HOME:-/tmp}"

dart analyze \
  lib/two_finance_blockchain.dart \
  lib/wallet.dart \
  lib/lifecycle.dart \
  lib/wallet_manager.dart \
  lib/blockchain/block/block.dart \
  test/helpers/fake_mqtt.dart \
  test/spec_harness/spec_files_test.dart \
  test/two_finance_blockchain_client_unit_test.dart \
  test/blockchain/contract/walletV2/wallet_unit_test.dart \
  test/blockchain/contract/lifecycle/lifecycle_test.dart \
  test/wallet_manager/wallet_manager_test.dart

dart test \
  test/spec_harness/spec_files_test.dart \
  test/two_finance_blockchain_client_unit_test.dart \
  test/blockchain/contract/walletV2/wallet_unit_test.dart \
  test/blockchain/contract/lifecycle/lifecycle_test.dart \
  test/wallet_manager/wallet_manager_test.dart

git diff --check
