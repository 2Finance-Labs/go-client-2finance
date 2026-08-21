# Quality Feedback

Este repo tem três velocidades de feedback:

## Fast

Roda sem broker, sem rede e sem infraestrutura externa.

```bash
bash tool/check.sh
```

Inclui:

- análise dos arquivos centrais do SDK e harness;
- validação de specs;
- testes unitários de contratos de payload;
- testes de wallet manager.

## Contract

Valida compatibilidade observável entre o client Dart e o contrato de rede/Go, principalmente métodos MQTT e payloads.

Arquivos principais:

- `specs/**/*.spec.json`
- `test/spec_harness/spec_files_test.dart`
- `test/two_finance_blockchain_client_unit_test.dart`
- `test/blockchain/contract/walletV2/wallet_unit_test.dart`
- `test/blockchain/contract/lifecycle/lifecycle_test.dart`

## E2E

Roda contra infraestrutura real. Deve provar que os fluxos principais ainda funcionam, mas não deve ser a primeira linha de feedback.

```bash
RUN_E2E_MQTT=1 MQTT_HOST=127.0.0.1 MQTT_PORT=1883 dart test -t e2e
```

## Checklist Antes De PR

- Specs atualizadas quando comportamento público muda.
- Harness de specs passando.
- Teste unitário de contrato para qualquer payload novo.
- E2E só quando o risco exige integração real.
- `git diff --check` limpo.
- Nenhum warning novo em `dart analyze` nos arquivos alterados.

## Próximos Checkers Úteis

- Comparador automático Dart vs Go para nomes de método e payloads críticos.
- Coverage mínimo só para harness/contratos pequenos, sem exigir coverage global ainda.
- Golden JSON para requests MQTT por contrato.
- CI separado para E2E manual ou agendado.
- Validador que exige que cada spec case tenha um teste Dart correspondente.
