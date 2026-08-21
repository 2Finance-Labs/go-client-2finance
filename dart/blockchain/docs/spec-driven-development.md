# Spec-Driven Development e Harness

Este repositório deve tratar especificações como artefatos versionados, revisáveis e executáveis. A proposta aqui não é substituir os testes Dart existentes, mas criar uma camada acima deles: cada comportamento importante ganha uma spec pequena, com exemplos concretos, e o harness garante que essas specs continuam válidas e rastreáveis.

## Referências

- Martin Fowler, Specification By Example: exemplos ajudam a especificar comportamento, mas não substituem conversa, design e outras técnicas de validação.
- Martin Fowler, Test Pyramid: priorizar muitos testes pequenos e rápidos, deixando testes E2E como segunda linha de defesa.
- Martin Fowler, Given When Then: estruturar cenários em precondição, ação e resultado esperado.
- Cucumber/Gherkin Reference: `Feature`, `Rule`, `Example`, `Given`, `When`, `Then`, tags e tabelas são úteis como linguagem, mesmo quando não usamos Cucumber como runner.
- Google Testing Blog, Test Sizes: separar testes pequenos, médios e grandes ajuda a manter feedback rápido e confiável.
- Google Testing Blog, Just Say No to More End-to-End Tests: testes E2E são valiosos, mas caros e frágeis quando viram a principal defesa.

## Decisão Para Este Repo

O projeto é um SDK Dart que fala com uma rede via MQTT e precisa ficar compatível com o cliente Go. Por isso, o harness mais valioso aqui é um harness de contrato de client:

1. Specs descrevem chamadas públicas do SDK, payload MQTT esperado e resposta esperada.
2. Testes unitários usam fakes, como `FakeMqttClient`, para validar payloads sem broker.
3. Testes E2E continuam existindo para provar integração real com a rede.
4. Specs ficam pequenas e próximas do domínio: wallet, token, lifecycle, raffle, cashback, client queries.

## Camadas

- `spec`: documento JSON versionado em `specs/`.
- `harness-small`: valida schema das specs e executa contratos sem rede.
- `contract-unit`: testes Dart com fake MQTT que validam payload/método.
- `e2e`: testes que falam com infraestrutura real.

## Convenções

- Um arquivo por área funcional: `wallet_v2.spec.json`, `lifecycle_v2.spec.json`.
- Cada spec tem `id`, `title`, `owner`, `status`, `layers` e `cases`.
- Cada case tem `given`, `when`, `then` e pelo menos um `tags`.
- `given` descreve estado e dados.
- `when` descreve a chamada pública do SDK.
- `then` descreve contrato observável: request method, transaction method, `to`, `data`, ou output.
- Não colocar detalhes irrelevantes de implementação nas specs.
- Quando um bug E2E aparecer, criar primeiro uma case pequena de contrato que reproduza o comportamento.

## Fluxo De Trabalho

1. Escrever ou atualizar spec antes da implementação.
2. Adicionar teste unitário/harness para a case nova.
3. Implementar o SDK.
4. Rodar `dart test test/spec_harness/spec_files_test.dart` e os testes da área.
5. Se necessário, promover uma case crítica para E2E.

## Estrutura

```text
docs/
  spec-driven-development.md
specs/
  README.md
  client/
    queries.spec.json
  contracts/
    lifecycle_v2.spec.json
    wallet_v2.spec.json
test/
  spec_harness/
    spec_files_test.dart
```

## O Que Não Fazer

- Não transformar todo teste em E2E.
- Não escrever specs gigantes com fluxo completo de produto quando uma regra pequena basta.
- Não duplicar steps textuais sem necessidade.
- Não validar detalhes internos que o usuário ou sistema externo não observa.
- Não deixar specs divergirem dos testes; se a spec muda, o harness/teste deve mudar junto.
