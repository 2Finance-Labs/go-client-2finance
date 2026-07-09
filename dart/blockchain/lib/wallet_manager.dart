import 'dart:async';
import 'dart:convert';
import 'dart:io';
import 'dart:math';
import 'dart:typed_data';

import 'package:cryptography/cryptography.dart';
import 'package:two_finance_blockchain/blockchain/keys/keys.dart';
import 'package:two_finance_blockchain/blockchain/transaction/transaction.dart';
import 'package:two_finance_blockchain/blockchain/utils/json.dart';

const int _walletVersion = 1;
const Duration _unlockDuration = Duration(minutes: 2);
const String _walletFileAssociatedData = 'wallet-manager-file:v1';
const String _privateKeyAssociatedDataPrefix = 'wallet-manager-private-key:';

enum WalletAction { exportPrivateKey, changePassword, deleteWallet, withdraw }

class KeysetKDFParams {
  final String alg;
  final int time;
  final int memoryKB;
  final int parallel;
  final int keyLen;
  final Uint8List salt;

  KeysetKDFParams({
    required this.alg,
    required this.time,
    required this.memoryKB,
    required this.parallel,
    required this.keyLen,
    required this.salt,
  });

  factory KeysetKDFParams.create() {
    return KeysetKDFParams(
      alg: 'argon2id',
      time: 3,
      memoryKB: 64 * 1024,
      parallel: 1,
      keyLen: 32,
      salt: _randomBytes(16),
    );
  }

  factory KeysetKDFParams.fromJson(Map<String, dynamic> json) {
    return KeysetKDFParams(
      alg: json['alg'] as String,
      time: json['time'] as int,
      memoryKB: json['memory_kb'] as int,
      parallel: json['parallel'] as int,
      keyLen: json['key_len'] as int,
      salt: _decodeBytes(json['salt']),
    );
  }

  Map<String, dynamic> toJson() => {
    'alg': alg,
    'time': time,
    'memory_kb': memoryKB,
    'parallel': parallel,
    'key_len': keyLen,
    'salt': base64Encode(salt),
  };
}

class LocalEncryptedWalletFile {
  final KeysetKDFParams kdf;
  final Uint8List cipher;

  LocalEncryptedWalletFile({required this.kdf, required this.cipher});

  factory LocalEncryptedWalletFile.fromJson(Map<String, dynamic> json) {
    return LocalEncryptedWalletFile(
      kdf: KeysetKDFParams.fromJson(json['kdf'] as Map<String, dynamic>),
      cipher: _decodeBytes(json['cipher']),
    );
  }

  Map<String, dynamic> toJson() => {
    'kdf': kdf.toJson(),
    'cipher': base64Encode(cipher),
  };
}

class WalletFile {
  final int version;
  final String owner;
  final Uint8List encryptedPrivateKey;
  final KeysetKDFParams privateKeyKdf;
  final DateTime createdAt;
  final DateTime updatedAt;

  WalletFile({
    required this.version,
    required this.owner,
    required this.encryptedPrivateKey,
    required this.privateKeyKdf,
    required this.createdAt,
    required this.updatedAt,
  });

  factory WalletFile.fromJson(Map<String, dynamic> json) {
    return WalletFile(
      version: json['version'] as int,
      owner: json['owner'] as String,
      encryptedPrivateKey: _decodeBytes(json['encrypted_private_key']),
      privateKeyKdf: KeysetKDFParams.fromJson(
        json['private_key_kdf'] as Map<String, dynamic>,
      ),
      createdAt: DateTime.parse(json['created_at'] as String),
      updatedAt: DateTime.parse(json['updated_at'] as String),
    );
  }

  Map<String, dynamic> toJson() => {
    'version': version,
    'owner': owner,
    'encrypted_private_key': base64Encode(encryptedPrivateKey),
    'private_key_kdf': privateKeyKdf.toJson(),
    'created_at': createdAt.toUtc().toIso8601String(),
    'updated_at': updatedAt.toUtc().toIso8601String(),
  };
}

class SignTransactionInput {
  final int chainID;
  final String from;
  final String to;
  final String method;
  final JsonMessage data;
  final int version;
  final String uuid7;
  final AuthorizationEnvelope? authorization;

  SignTransactionInput({
    required this.chainID,
    required this.from,
    required this.to,
    required this.method,
    required this.data,
    required this.version,
    required this.uuid7,
    this.authorization,
  });
}

class PreparedTransaction {
  final int chainID;
  final String from;
  final String to;
  final String method;
  final JsonMessage data;
  final int version;
  final String uuid7;
  final String hash;
  final String signature;
  final AuthorizationEnvelope? authorization;

  PreparedTransaction({
    required this.chainID,
    required this.from,
    required this.to,
    required this.method,
    required this.data,
    required this.version,
    required this.uuid7,
    this.hash = '',
    this.signature = '',
    this.authorization,
  });

  factory PreparedTransaction.fromJson(Map<String, dynamic> json) {
    return PreparedTransaction(
      chainID: json['chain_id'] as int,
      from: json['from'] as String,
      to: json['to'] as String,
      method: json['method'] as String,
      data: Map<String, dynamic>.from(json['data'] as Map),
      version: json['version'] as int,
      uuid7: json['uuid7'] as String,
      hash: json['hash'] as String? ?? '',
      signature: json['signature'] as String? ?? '',
      authorization: json['authorization'] == null
          ? null
          : AuthorizationEnvelope.fromJson(
              Map<String, dynamic>.from(json['authorization'] as Map),
            ),
    );
  }

  Map<String, dynamic> toJson() => {
    'chain_id': chainID,
    'from': from,
    'to': to,
    'method': method,
    'data': data,
    'version': version,
    'uuid7': uuid7,
    'hash': hash,
    'signature': signature,
    if (authorization != null) 'authorization': authorization!.toJson(),
  };
}

class SignedTransaction extends PreparedTransaction {
  SignedTransaction({
    required super.chainID,
    required super.from,
    required super.to,
    required super.method,
    required super.data,
    required super.version,
    required super.uuid7,
    required super.hash,
    required super.signature,
    super.authorization,
  });

  factory SignedTransaction.fromTransaction(Transaction tx) {
    return SignedTransaction(
      chainID: tx.chainID,
      from: tx.from,
      to: tx.to,
      method: tx.method,
      data: tx.data,
      version: tx.version,
      uuid7: tx.uuid7,
      hash: tx.hash,
      signature: tx.signature,
      authorization: tx.authorization,
    );
  }

  factory SignedTransaction.fromJson(Map<String, dynamic> json) {
    return SignedTransaction(
      chainID: json['chain_id'] as int,
      from: json['from'] as String,
      to: json['to'] as String,
      method: json['method'] as String,
      data: Map<String, dynamic>.from(json['data'] as Map),
      version: json['version'] as int,
      uuid7: json['uuid7'] as String,
      hash: json['hash'] as String,
      signature: json['signature'] as String,
      authorization: json['authorization'] == null
          ? null
          : AuthorizationEnvelope.fromJson(
              Map<String, dynamic>.from(json['authorization'] as Map),
            ),
    );
  }

  Transaction toNetworkTransaction() {
    return Transaction(
      chainID: chainID,
      from: from,
      to: to,
      method: method,
      data: data,
      version: version,
      uuid7: uuid7,
      hash: hash,
      signature: signature,
      authorization: authorization,
    );
  }
}

abstract class IWalletManager {
  Future<void> importPrivateKey(Uint8List privateKey, String password);
  Future<void> lock();
  Future<void> unlockWithPassword(String password);
  bool isUnlocked();
  Future<void> changePassword(String currentPassword, String newPassword);
  String get ownerAddress;
  Future<Transaction> signTransaction(SignTransactionInput input);
  Future<SignedTransaction> signPreparedTransaction(PreparedTransaction input);
}

class WalletManager implements IWalletManager {
  final String filePath;
  String _owner = '';
  Uint8List? _privateKey;
  DateTime? _unlockedUntil;
  Timer? _lockTimer;

  WalletManager(this.filePath);

  @override
  String get ownerAddress => _owner;

  @override
  Future<void> importPrivateKey(Uint8List privateKey, String password) async {
    if (password.isEmpty) {
      throw ArgumentError('password is required');
    }
    if (privateKey.isEmpty) {
      throw ArgumentError('private key is required');
    }
    if (filePath.isEmpty) {
      throw StateError('wallet file path is required');
    }

    final privateKeyHex = utf8.decode(privateKey);
    final owner = await _publicKeyFromPrivateHex(privateKeyHex);
    KeyManager.validateEDDSAPublicKeyHex(owner);

    final privateKeyKdf = KeysetKDFParams.create();
    final encryptedPrivateKey = await _encryptWithPassword(
      privateKey,
      password,
      privateKeyKdf,
      utf8.encode('$_privateKeyAssociatedDataPrefix$owner'),
    );

    final now = DateTime.now().toUtc();
    final walletFile = WalletFile(
      version: _walletVersion,
      owner: owner,
      encryptedPrivateKey: encryptedPrivateKey,
      privateKeyKdf: privateKeyKdf,
      createdAt: now,
      updatedAt: now,
    );

    final payload = utf8.encode(jsonEncode(walletFile.toJson()));
    final localFile = await _encryptWalletPayload(payload, password);
    await _writeEncryptedWalletFile(localFile);

    _owner = owner;
    _clearBytes(privateKey);
    await lock();
  }

  @override
  Future<void> lock() async {
    _lockTimer?.cancel();
    _lockTimer = null;
    final privateKey = _privateKey;
    if (privateKey != null) {
      _clearBytes(privateKey);
    }
    _privateKey = null;
    _unlockedUntil = null;
  }

  @override
  Future<void> unlockWithPassword(String password) async {
    if (password.isEmpty) {
      throw ArgumentError('password is required');
    }
    if (filePath.isEmpty) {
      throw StateError('wallet file path is required');
    }

    final walletFile = await _readWalletFile(password);
    if (_owner.isNotEmpty && walletFile.owner != _owner) {
      throw StateError('wallet owner mismatch');
    }

    final privateKey = await _decryptWithPassword(
      walletFile.encryptedPrivateKey,
      password,
      walletFile.privateKeyKdf,
      utf8.encode('$_privateKeyAssociatedDataPrefix${walletFile.owner}'),
    );

    await lock();
    _owner = walletFile.owner;
    _privateKey = Uint8List.fromList(privateKey);
    _unlockedUntil = DateTime.now().add(_unlockDuration);
    _lockTimer = Timer(_unlockDuration, () {
      unawaited(lock());
    });
  }

  @override
  bool isUnlocked() {
    final privateKey = _privateKey;
    final unlockedUntil = _unlockedUntil;
    return privateKey != null &&
        privateKey.isNotEmpty &&
        unlockedUntil != null &&
        DateTime.now().isBefore(unlockedUntil);
  }

  @override
  Future<void> changePassword(
    String currentPassword,
    String newPassword,
  ) async {
    if (currentPassword.isEmpty) {
      throw ArgumentError('current password is required');
    }
    if (newPassword.isEmpty) {
      throw ArgumentError('new password is required');
    }
    if (currentPassword == newPassword) {
      throw ArgumentError(
        'new password must be different from current password',
      );
    }
    if (_owner.isEmpty) {
      throw StateError('owner is required');
    }

    final walletFile = await _readWalletFile(currentPassword);
    if (walletFile.owner != _owner) {
      throw StateError('wallet owner mismatch');
    }

    final privateKey = await _decryptWithPassword(
      walletFile.encryptedPrivateKey,
      currentPassword,
      walletFile.privateKeyKdf,
      utf8.encode('$_privateKeyAssociatedDataPrefix${walletFile.owner}'),
    );

    final newPrivateKeyKdf = KeysetKDFParams.create();
    final encryptedPrivateKey = await _encryptWithPassword(
      privateKey,
      newPassword,
      newPrivateKeyKdf,
      utf8.encode('$_privateKeyAssociatedDataPrefix${walletFile.owner}'),
    );

    final updatedWalletFile = WalletFile(
      version: walletFile.version,
      owner: walletFile.owner,
      encryptedPrivateKey: encryptedPrivateKey,
      privateKeyKdf: newPrivateKeyKdf,
      createdAt: walletFile.createdAt,
      updatedAt: DateTime.now().toUtc(),
    );

    final payload = utf8.encode(jsonEncode(updatedWalletFile.toJson()));
    final localFile = await _encryptWalletPayload(payload, newPassword);
    await _writeEncryptedWalletFile(localFile);
    _clearBytes(Uint8List.fromList(privateKey));
    await lock();
  }

  @override
  Future<Transaction> signTransaction(SignTransactionInput input) async {
    if (input.method.isEmpty) {
      throw ArgumentError('method is required');
    }
    if (!isUnlocked()) {
      throw StateError('wallet is locked');
    }

    final privateKey = Uint8List.fromList(_privateKey!);
    try {
      final tx = Transaction.create(
        chainID: input.chainID,
        from: input.from,
        to: input.to,
        method: input.method,
        data: input.data,
        version: input.version,
        uuid7: input.uuid7,
        authorization: input.authorization,
      );
      return signTransactionWithPrivateKey(privateKey, tx);
    } finally {
      _clearBytes(privateKey);
    }
  }

  @override
  Future<SignedTransaction> signPreparedTransaction(
    PreparedTransaction input,
  ) async {
    final signed = await signTransaction(
      SignTransactionInput(
        chainID: input.chainID,
        from: input.from,
        to: input.to,
        method: input.method,
        data: input.data,
        version: input.version,
        uuid7: input.uuid7,
        authorization: input.authorization,
      ),
    );
    return SignedTransaction.fromTransaction(signed);
  }

  Future<WalletFile> _readWalletFile(String password) async {
    final encryptedFile = File(filePath);
    if (!encryptedFile.existsSync()) {
      throw StateError('failed to read wallet file');
    }

    final localFile = LocalEncryptedWalletFile.fromJson(
      jsonDecode(await encryptedFile.readAsString()) as Map<String, dynamic>,
    );
    final payload = await _decryptWithPassword(
      localFile.cipher,
      password,
      localFile.kdf,
      utf8.encode(_walletFileAssociatedData),
    );
    final walletFile = WalletFile.fromJson(
      jsonDecode(utf8.decode(payload)) as Map<String, dynamic>,
    );

    if (walletFile.version != _walletVersion) {
      throw StateError('unsupported wallet version: ${walletFile.version}');
    }
    if (walletFile.owner.isEmpty) {
      throw StateError('wallet owner is required');
    }
    if (walletFile.encryptedPrivateKey.isEmpty) {
      throw StateError('encrypted private key is required');
    }
    return walletFile;
  }

  Future<LocalEncryptedWalletFile> _encryptWalletPayload(
    List<int> payload,
    String password,
  ) async {
    final kdf = KeysetKDFParams.create();
    final cipher = await _encryptWithPassword(
      payload,
      password,
      kdf,
      utf8.encode(_walletFileAssociatedData),
    );
    return LocalEncryptedWalletFile(kdf: kdf, cipher: cipher);
  }

  Future<void> _writeEncryptedWalletFile(
    LocalEncryptedWalletFile localFile,
  ) async {
    final walletFile = File(filePath);
    await walletFile.parent.create(recursive: true);
    await walletFile.writeAsString(jsonEncode(localFile.toJson()));
  }
}

Future<Transaction> signTransactionWithPrivateKey(
  Uint8List privateKey,
  Transaction tx,
) {
  return signTransaction(utf8.decode(privateKey), tx);
}

Future<KeyPair2Finance> generateEd25519KeyPairHex() {
  return KeyManager().generateKeyEd25519();
}

Future<String> _publicKeyFromPrivateHex(String privateKeyHex) async {
  final privateKeyBytes = KeyManager.hexToBytes(privateKeyHex);
  if (privateKeyBytes.length < 32) {
    throw ArgumentError('private key must be at least 32 bytes');
  }
  final keyPair = await Ed25519().newKeyPairFromSeed(
    privateKeyBytes.sublist(0, 32),
  );
  final publicKey = await keyPair.extractPublicKey();
  return KeyManager.bytesToHex(publicKey.bytes);
}

Future<Uint8List> _encryptWithPassword(
  List<int> data,
  String password,
  KeysetKDFParams params,
  List<int> associatedData,
) async {
  final key = await _derivePasswordKey(password, params);
  final box = await Xchacha20.poly1305Aead().encrypt(
    data,
    secretKey: key,
    aad: associatedData,
  );
  return Uint8List.fromList(box.concatenation());
}

Future<List<int>> _decryptWithPassword(
  List<int> cipher,
  String password,
  KeysetKDFParams params,
  List<int> associatedData,
) async {
  final key = await _derivePasswordKey(password, params);
  final box = SecretBox.fromConcatenation(
    cipher,
    nonceLength: Xchacha20.poly1305Aead().nonceLength,
    macLength: 16,
  );
  try {
    return await Xchacha20.poly1305Aead().decrypt(
      box,
      secretKey: key,
      aad: associatedData,
    );
  } on SecretBoxAuthenticationError {
    throw StateError('invalid password or corrupted keyset');
  }
}

Future<SecretKey> _derivePasswordKey(String password, KeysetKDFParams params) {
  if (password.isEmpty) {
    throw ArgumentError('password is required');
  }
  if (params.alg != 'argon2id') {
    throw ArgumentError('unsupported KDF algorithm: ${params.alg}');
  }
  if (params.salt.isEmpty) {
    throw ArgumentError('salt is required');
  }
  return Argon2id(
    parallelism: params.parallel,
    memory: params.memoryKB,
    iterations: params.time,
    hashLength: params.keyLen,
  ).deriveKey(secretKey: SecretKey(utf8.encode(password)), nonce: params.salt);
}

Uint8List _randomBytes(int length) {
  final random = Random.secure();
  return Uint8List.fromList(
    List<int>.generate(length, (_) => random.nextInt(256)),
  );
}

Uint8List _decodeBytes(dynamic value) {
  if (value is String) {
    return Uint8List.fromList(base64Decode(value));
  }
  if (value is List) {
    return Uint8List.fromList(value.cast<int>());
  }
  throw ArgumentError('invalid byte encoding');
}

void _clearBytes(Uint8List data) {
  for (var i = 0; i < data.length; i++) {
    data[i] = 0;
  }
}
