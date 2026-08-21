import 'package:flutter/material.dart';
import 'dart:async';
import 'dart:io';
import 'dart:math';

import 'package:flutter/services.dart';
import 'package:two_finance_blockchain/two_finance_blockchain.dart';
import 'package:two_finance_blockchain/infra/mqtt/mqtt.dart';
import 'package:two_finance_blockchain/config/config.dart';
import 'package:mqtt_client/mqtt_client.dart' show MqttClient, MqttConnectionState, MqttPublishMessage, MqttPublishPayload;
import 'package:mqtt_client/mqtt_server_client.dart';
import 'package:two_finance_blockchain/blockchain/keys/keys.dart';
import 'package:two_finance_blockchain/blockchain/transaction/transaction.dart';
import 'package:two_finance_blockchain/blockchain/types/types.dart';
import 'package:two_finance_blockchain/blockchain/contract/walletV2/constants.dart';
import 'package:two_finance_blockchain/blockchain/contract/tokenV2/domain/token.dart' as domain;

import 'package:two_finance_blockchain/blockchain/contract/walletV2/domain/wallet.dart' as domain;


void main() {
  runApp(const MyApp());
}

class MyApp extends StatefulWidget {
  const MyApp({super.key});

  @override
  State<MyApp> createState() => _MyAppState();
}

class _MyAppState extends State<MyApp> {
  String _platformVersion = 'Unknown';
  // Instância do plugin TwoFinanceBlockchain
  late MqttClientWrapper mqttClient;
  late TwoFinanceBlockchain _twoFinanceBlockchainPlugin;
  late KeyManager keyManager;
  

  @override
  void initState() {
    super.initState();
    initPlatformState();
  }

  // Platform messages are asynchronous, so we initialize in an async method.
  Future<void> initPlatformState() async {
    String platformVersion;

    _twoFinanceBlockchainPlugin = await initBlockchain();
    // Platform messages may fail, so we use a try/catch PlatformException.
    // We also handle the message potentially returning null.
    try {
      platformVersion =
          await _twoFinanceBlockchainPlugin.getPlatformVersion() ?? 'Unknown platform version';
    } on PlatformException {
      platformVersion = 'Failed to get platform version.';
    }

    // If the widget was removed from the tree while the asynchronous platform
    // message was in flight, we want to discard the reply rather than calling
    // setState to update our non-existent appearance.
    if (!mounted) return;

    setState(() {
      _platformVersion = platformVersion;
    });
  }

  Future<void> testMqttClient() async {
  try {

    print('🔌 Connecting...');
    await mqttClient.connect();

    print('📡 Subscribing...');
    await mqttClient.subscribe('test/topic', handler: (mqttClient, message) {
      final mqttMessage = message.payload as MqttPublishMessage;
      final payload = MqttPublishPayload.bytesToStringAsString(mqttMessage.payload.message);
      print('📥 Message received on ${message.topic}: $payload');
    });

    print('📤 Publishing...');
    await mqttClient.publish('test/topic', 'Hello from Flutter MQTT!');

    //await Future.delayed(const Duration(seconds: 3));

    print('🚫 Unsubscribing...');
    await mqttClient.unsubscribe('test/topic');

    print('🔌 Disconnecting...');
    await mqttClient.disconnect();
    print('✅ MQTT Client test completed successfully.');
    } catch (e) {
      print('❌ Error: $e');
      rethrow; // throws the error up to the caller
    }
  }

  String generateRandomSuffix(int length) {
    const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789';
    final rand = Random.secure();
    return List.generate(length, (_) => chars[rand.nextInt(chars.length)]).join();
  }

  Future<TwoFinanceBlockchain> initBlockchain() async {
    try {

      await Config.loadConfig(env: 'prod', path: 'packages/two_finance_blockchain/assets/.env');
      
      mqttClient = MqttClientWrapper(
        host: Config.emqxHost,
        port: Config.emqxPort,
        clientId: Config.emqxClientId,
        useSSL: Config.emqxSSL,
        username: Config.emqxUsername,
        password: Config.emqxPassword,
        caCertPath: Config.emqxCaCertPath,
      );
      await mqttClient.connect();
      print('Initializing TwoFinanceBlockchain plugin...');
      
      keyManager = KeyManager();
      
      // ✅ Correct assignment to class-level field
      final plugin = TwoFinanceBlockchain(
        keyManager: keyManager,
        mqttClient: mqttClient,
      );

      return plugin;
    } catch (e) {
      print('⚠️ Error initializing TwoFinanceBlockchain: $e');
      rethrow;
    }
  }

  Future<void> testBlockchain() async {
    print('Plugin initialized successfully.');
    await _twoFinanceBlockchainPlugin.initialize();
    
    final keyPair1 = await _twoFinanceBlockchainPlugin.generateKeyEd25519();
    final keyPair2 = await _twoFinanceBlockchainPlugin.generateKeyEd25519();

    await _twoFinanceBlockchainPlugin.setPrivateKey(keyPair1.privateKey);
    final publicKey1 = keyPair1.publicKey;
    final publicKey2 = keyPair2.publicKey;
    print('Private Key: ${keyPair1.privateKey}');
    print('Public Key: $publicKey1');
    print('Receiver Public Key: $publicKey2');
    
  }

 @override
  Widget build(BuildContext context) {
    return MaterialApp(
      home: Scaffold(
        appBar: AppBar(
          title: const Text('Plugin example app'),
        ),
        body: Center(
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Text('Running on: $_platformVersion\n'),
              const SizedBox(height: 20),
              ElevatedButton(
                onPressed: testBlockchain,
                child: const Text('Run Blockchain Test'),
              ),
            ],
          ),
        ),
      ),
    );
}


  
}
