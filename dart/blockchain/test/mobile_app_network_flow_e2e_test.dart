import 'dart:io';

import 'package:test/test.dart';
import 'package:two_finance_blockchain/two_finance_blockchain.dart';

import 'helpers/helpers.dart';

const _fxLifecycleV2 = 'fxLifecycleV2';
const _receivingLifecycleV2 = 'receivingLifecycleV2';
const _sendingLifecycleV2 = 'sendingLifecycleV2';
const _multiCurrencyLifecycleV2 = 'multiCurrencyLifecycleV2';

void main() {
  final runE2E = Platform.environment['RUN_E2E_MQTT'] == '1';

  group(
    'Mobile app network flows E2E',
    skip: runE2E ? null : 'Set RUN_E2E_MQTT=1 to run against 2finance-network.',
    () {
      test(
        'starts mobile receive, send, pix, fx and multi-currency lifecycles on 2finance-network',
        () async {
          final client = await setupClient();
          final signer = await validKeyPair();
          await client.setPrivateKey(signer.privateKey);

          final fxAddress = await _deploy(client, _fxLifecycleV2);
          final receivingAddress = await _deploy(client, _receivingLifecycleV2);
          final sendingAddress = await _deploy(client, _sendingLifecycleV2);
          final multiAddress = await _deploy(client, _multiCurrencyLifecycleV2);

          final receiver = await validKeyPair();
          final suffix = DateTime.now().microsecondsSinceEpoch;

          final receivingRequestId = 'mobile-receive-$suffix';
          final receiving = await client.startReceiving(
            address: receivingAddress,
            data: {
              'request_id': receivingRequestId,
              'owner': signer.publicKey,
              'provider': 'codexa',
              'provider_account_id': 'acct-mobile-e2e',
              'external_auth_id': 'auth-mobile-e2e',
              'profile_id': 'profile-mobile-e2e',
              'currency': 'BRL',
              'amount': '25.50',
              'details': {
                'payment_method': 'pix',
                'receiver_public_key': signer.publicKey,
              },
            },
          );
          _expectLog(receiving, 'ReceivingLifecycle_Started', receivingAddress);
          await _expectState(
            () => client.getReceiving(
              address: receivingAddress,
              requestId: receivingRequestId,
            ),
            receivingRequestId,
          );

          final sendingRequestId = 'mobile-send-$suffix';
          final sending = await client.startSending(
            address: sendingAddress,
            data: {
              'request_id': sendingRequestId,
              'owner': signer.publicKey,
              'provider': 'codexa',
              'provider_account_id': 'acct-mobile-e2e',
              'external_auth_id': 'auth-mobile-e2e',
              'profile_id': 'profile-mobile-e2e',
              'source_currency': 'BRL',
              'target_currency': 'USD',
              'source_wallet_id': 'wallet-mobile-e2e',
              'beneficiary_name': 'John Smith',
              'beneficiary_type': 'individual',
              'beneficiary_country': 'US',
              'currency': 'USD',
              'amount': '10.00',
              'reference': sendingRequestId,
              'reason': 'mobile_send_e2e',
              'details': {'receiver_public_key': receiver.publicKey},
            },
          );
          _expectLog(sending, 'SendingLifecycle_Started', sendingAddress);
          await _expectState(
            () => client.getSending(
              address: sendingAddress,
              requestId: sendingRequestId,
            ),
            sendingRequestId,
          );

          final pixTopupRequestId = 'mobile-pixin-$suffix';
          final pixTopup = await client.startReceiving(
            address: receivingAddress,
            data: {
              'request_id': pixTopupRequestId,
              'owner': signer.publicKey,
              'provider': 'wise',
              'provider_account_id': 'acct-mobile-e2e',
              'external_auth_id': 'auth-mobile-e2e',
              'profile_id': 'profile-mobile-e2e',
              'currency': 'BRL',
              'amount': '25.50',
              'reference': pixTopupRequestId,
              'reason': 'mobile_pix_topup',
              'details': {
                'payment_method': 'pix',
                'owner_public_key': signer.publicKey,
              },
            },
          );
          _expectLog(pixTopup, 'ReceivingLifecycle_Started', receivingAddress);
          await _expectState(
            () => client.getReceiving(
              address: receivingAddress,
              requestId: pixTopupRequestId,
            ),
            pixTopupRequestId,
          );

          final pixRescueRequestId = 'mobile-pixout-$suffix';
          final pixRescue = await client.startSending(
            address: sendingAddress,
            data: {
              'request_id': pixRescueRequestId,
              'owner': signer.publicKey,
              'provider': 'wise',
              'provider_account_id': 'acct-mobile-e2e',
              'external_auth_id': 'auth-mobile-e2e',
              'profile_id': 'profile-mobile-e2e',
              'currency': 'BRL',
              'amount': '12.30',
              'reference': pixRescueRequestId,
              'reason': 'mobile_pix_rescue',
              'receiver_public_key': '11111124591',
              'details': {
                'payment_method': 'pix',
                'pix_key': '11111124591',
                'destination_name': 'Aline Dias',
                'destination_document': '11111124591',
                'destination_institution': 'Nu Pagamentos',
              },
            },
          );
          _expectLog(pixRescue, 'SendingLifecycle_Started', sendingAddress);
          await _expectState(
            () => client.getSending(
              address: sendingAddress,
              requestId: pixRescueRequestId,
            ),
            pixRescueRequestId,
          );

          final fxRequestId = 'mobile-fx-$suffix';
          final fx = await client.startFX(
            address: fxAddress,
            data: {
              'request_id': fxRequestId,
              'owner': signer.publicKey,
              'provider': 'wise',
              'provider_account_id': 'acct-mobile-e2e',
              'external_auth_id': 'auth-mobile-e2e',
              'profile_id': 'profile-mobile-e2e',
              'source_currency': 'BRL',
              'target_currency': 'USD',
              'source_amount': '100.00',
              'target_amount': '18.00',
              'reference': fxRequestId,
              'reason': 'mobile_fx_e2e',
              'details': {'source': 'mobile_app'},
            },
          );
          _expectLog(fx, 'FXLifecycle_Started', fxAddress);
          await _expectState(
            () => client.getFX(address: fxAddress, requestId: fxRequestId),
            fxRequestId,
          );

          final multiRequestId = 'mobile-multi-$suffix';
          final multi = await client.startMultiCurrency(
            address: multiAddress,
            data: {
              'request_id': multiRequestId,
              'owner': signer.publicKey,
              'provider': 'wise',
              'provider_account_id': 'acct-mobile-e2e',
              'external_auth_id': 'auth-mobile-e2e',
              'profile_id': 'profile-mobile-e2e',
              'currency': 'USD',
              'details': {
                'currencies': ['BRL', 'USD', 'EUR'],
                'owner_public_key': signer.publicKey,
              },
            },
          );
          _expectLog(multi, 'MultiCurrencyLifecycle_Started', multiAddress);
          await _expectState(
            () => client.getMultiCurrency(
              address: multiAddress,
              requestId: multiRequestId,
            ),
            multiRequestId,
          );
        },
        timeout: const Timeout(Duration(minutes: 3)),
        tags: const ['e2e'],
      );
    },
  );
}

Future<String> _deploy(TwoFinanceBlockchain client, String version) async {
  final output = await client.deployContract1(version);
  expect(output.logs, isNotNull);
  expect(output.logs, isNotEmpty);
  return output.logs!.first.contractAddress;
}

void _expectLog(
  ContractOutput output,
  String expectedLogType,
  String expectedAddress,
) {
  expect(output.logs, isNotNull);
  expect(output.logs, isNotEmpty);
  expect(output.logs!.first.logType, expectedLogType);
  expect(output.logs!.first.contractAddress, expectedAddress);
}

Future<void> _expectState(
  Future<ContractOutput> Function() read,
  String requestId,
) async {
  ContractOutput? output;
  for (var attempt = 0; attempt < 10; attempt++) {
    output = await read();
    if (output.states != null && output.states!.isNotEmpty) {
      break;
    }
    await Future<void>.delayed(const Duration(milliseconds: 250));
  }

  expect(output, isNotNull);
  expect(output!.states, isNotNull);
  expect(output.states, isNotEmpty);
  final state = Map<String, dynamic>.from(output.states!.first.object as Map);
  expect(state['request_id'], requestId);
}
