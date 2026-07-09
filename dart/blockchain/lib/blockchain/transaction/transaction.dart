import 'dart:convert';
import 'dart:typed_data';
import 'package:cryptography/cryptography.dart';
import 'package:web3dart/crypto.dart';
import 'package:two_finance_blockchain/blockchain/keys/keys.dart';
import 'package:two_finance_blockchain/blockchain/types/types.dart';
import 'package:two_finance_blockchain/blockchain/utils/json.dart';
import 'package:two_finance_blockchain/blockchain/utils/uuid.dart';

abstract class ITransaction {
  Future<void> validateUnsignedTransaction();
  Future<void> validateTransaction();
  Future<void> validateHash();
  Future<String> calculateHash();
  Transaction get();
}

class AuthorizationEnvelope {
  String type;
  int policyVersion;
  String nonceKey;
  int sequence;
  String? validUntil;
  AuthorizationPolicy? policy;
  List<AuthorizationSignature> userSignatures;
  DeviceAuthorization? device;
  List<AuthorizationSignature> keystoreSignatures;

  AuthorizationEnvelope({
    required this.type,
    required this.policyVersion,
    required this.nonceKey,
    required this.sequence,
    this.validUntil,
    this.policy,
    List<AuthorizationSignature>? userSignatures,
    this.device,
    List<AuthorizationSignature>? keystoreSignatures,
  }) : userSignatures = userSignatures ?? <AuthorizationSignature>[],
       keystoreSignatures = keystoreSignatures ?? <AuthorizationSignature>[];

  AuthorizationEnvelope withoutProofs() => AuthorizationEnvelope(
    type: type,
    policyVersion: policyVersion,
    nonceKey: nonceKey,
    sequence: sequence,
    validUntil: validUntil,
    policy: policy,
  );

  Map<String, dynamic> toJson() => {
    'type': type,
    'policy_version': policyVersion,
    'nonce_key': nonceKey,
    'sequence': sequence,
    if (validUntil != null) 'valid_until': validUntil,
    if (policy != null) 'policy': policy!.toJson(),
    if (userSignatures.isNotEmpty)
      'user_signatures': userSignatures.map((s) => s.toJson()).toList(),
    if (device != null) 'device': device!.toJson(),
    if (keystoreSignatures.isNotEmpty)
      'keystore_signatures': keystoreSignatures.map((s) => s.toJson()).toList(),
  };

  factory AuthorizationEnvelope.fromJson(
    Map<String, dynamic> json,
  ) => AuthorizationEnvelope(
    type: json['type'] as String,
    policyVersion: json['policy_version'] as int,
    nonceKey: json['nonce_key'] as String,
    sequence: json['sequence'] as int,
    validUntil: json['valid_until'] as String?,
    policy: json['policy'] == null
        ? null
        : AuthorizationPolicy.fromJson(
            Map<String, dynamic>.from(json['policy'] as Map),
          ),
    userSignatures: ((json['user_signatures'] as List?) ?? const [])
        .map(
          (s) => AuthorizationSignature.fromJson(
            Map<String, dynamic>.from(s as Map),
          ),
        )
        .toList(),
    device: json['device'] == null
        ? null
        : DeviceAuthorization.fromJson(
            Map<String, dynamic>.from(json['device'] as Map),
          ),
    keystoreSignatures: ((json['keystore_signatures'] as List?) ?? const [])
        .map(
          (s) => AuthorizationSignature.fromJson(
            Map<String, dynamic>.from(s as Map),
          ),
        )
        .toList(),
  );
}

class AuthorizationPolicy {
  String? accountID;
  int quorum;
  bool deviceRequired;
  List<PolicySigner> signers;

  AuthorizationPolicy({
    this.accountID,
    required this.quorum,
    required this.deviceRequired,
    required this.signers,
  });

  Map<String, dynamic> toJson() => {
    if (accountID != null && accountID!.isNotEmpty) 'account_id': accountID,
    'quorum': quorum,
    'device_required': deviceRequired,
    'signers': signers.map((s) => s.toJson()).toList(),
  };

  factory AuthorizationPolicy.fromJson(Map<String, dynamic> json) =>
      AuthorizationPolicy(
        accountID: json['account_id'] as String?,
        quorum: json['quorum'] as int,
        deviceRequired: json['device_required'] as bool? ?? false,
        signers: ((json['signers'] as List?) ?? const [])
            .map(
              (s) => PolicySigner.fromJson(Map<String, dynamic>.from(s as Map)),
            )
            .toList(),
      );
}

class PolicySigner {
  String publicKey;
  String kind;
  int weight;

  PolicySigner({
    required this.publicKey,
    required this.kind,
    required this.weight,
  });

  Map<String, dynamic> toJson() => {
    'public_key': publicKey,
    'kind': kind,
    'weight': weight,
  };

  factory PolicySigner.fromJson(Map<String, dynamic> json) => PolicySigner(
    publicKey: json['public_key'] as String,
    kind: json['kind'] as String,
    weight: json['weight'] as int,
  );
}

class AuthorizationSignature {
  String publicKey;
  String signature;

  AuthorizationSignature({required this.publicKey, required this.signature});

  Map<String, dynamic> toJson() => {
    'public_key': publicKey,
    'signature': signature,
  };

  factory AuthorizationSignature.fromJson(Map<String, dynamic> json) =>
      AuthorizationSignature(
        publicKey: json['public_key'] as String,
        signature: json['signature'] as String,
      );
}

class DeviceAuthorization {
  String deviceID;
  String publicKey;
  String signature;

  DeviceAuthorization({
    required this.deviceID,
    required this.publicKey,
    required this.signature,
  });

  Map<String, dynamic> toJson() => {
    'device_id': deviceID,
    'public_key': publicKey,
    'signature': signature,
  };

  factory DeviceAuthorization.fromJson(Map<String, dynamic> json) =>
      DeviceAuthorization(
        deviceID: json['device_id'] as String,
        publicKey: json['public_key'] as String,
        signature: json['signature'] as String,
      );
}

String derivePolicyAccountID(AuthorizationPolicy policy) {
  final normalized = AuthorizationPolicy(
    quorum: policy.quorum,
    deviceRequired: policy.deviceRequired,
    signers: [...policy.signers]
      ..sort((a, b) => a.publicKey.compareTo(b.publicKey)),
  );
  final canonical = canonicalJsonEncode({
    'domain': '2finance.network.account.policy.v1',
    'policy': normalized.toJson(),
  });
  return KeyManager.bytesToHex(
    keccak256(Uint8List.fromList(utf8.encode(canonical))),
  );
}

class Transaction implements ITransaction {
  int chainID;
  String from;
  String to;
  String method;
  JsonMessage data;
  int version;
  String uuid7;
  String hash;
  String signature;
  AuthorizationEnvelope? authorization;

  Transaction({
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

  @override
  String toString() {
    return '''
  Transaction(
    chainID: $chainID,
    from: $from,
    to: $to,
    method: $method,
    data: $data,
    version: $version,
    uuid7: $uuid7,
    hash: $hash,
    signature: $signature,
    authorization: $authorization,
  )
  ''';
  }

  static Transaction create({
    required int chainID,
    required String from,
    required String to,
    required String method,
    required JsonMessage data,
    required int version,
    required String uuid7,
    AuthorizationEnvelope? authorization,
  }) {
    return Transaction(
      chainID: chainID,
      from: from,
      to: to,
      method: method,
      data: data,
      version: version,
      uuid7: uuid7,
      authorization: authorization,
    );
  }

  @override
  Future<void> validateUnsignedTransaction() async {
    if (data.isEmpty) {
      throw Exception("data cannot be empty");
    }
  }

  @override
  Future<void> validateTransaction() async {
    if (chainID <= 0) {
      throw Exception("chain ID must be greater than zero");
    }
    if (chainID > 2) {
      throw Exception("unsupported chain ID");
    }
    if (from.isEmpty) {
      throw Exception("sender address is required");
    }
    if (to.isEmpty) {
      throw Exception("recipient address is required");
    }
    if (from == to) {
      throw Exception("sender and recipient cannot be the same");
    }
    if (method.isEmpty) {
      throw Exception("method is required");
    }
    if (data.isEmpty) {
      throw Exception("data cannot be empty");
    }
    if (version == 0) {
      throw Exception("version must be greater than zero");
    }

    if (authorization == null) {
      try {
        KeyManager.validateEDDSAPublicKeyHex(from);
      } catch (e) {
        throw Exception("invalid sender public key: $e");
      }
      if (signature.isEmpty) {
        throw Exception("signature cannot be empty");
      }
      if (signature.length != 128) {
        throw Exception("signature must be 128 characters long");
      }
    } else if (from.length != 64) {
      throw Exception("account id must be 64 hex characters long");
    }

    // Recipient pubkey validation (skip deploy address)
    if (to.isNotEmpty && to != DEPLOY_CONTRACT_ADDRESS) {
      try {
        KeyManager.validateEDDSAPublicKeyHex(to);
      } catch (e) {
        throw Exception("invalid recipient public key: $e");
      }
    }

    // UUIDv7 validation
    try {
      validateUUID7(uuid7);
    } catch (e) {
      throw Exception("invalid UUIDv7: $e");
    }

    if (hash.length != 64) {
      throw Exception("hash must be 64 characters long");
    }
    // Hash validation
    try {
      await validateHash();
    } catch (e) {
      throw Exception("transaction hash validation failed: $e");
    }
  }

  @override
  Future<void> validateHash() async {
    final computed = await calculateHash();
    if (computed != hash) {
      throw Exception("invalid hash: expected $computed, got $hash");
    }
  }

  @override
  Future<String> calculateHash() async {
    final temp = toJson();
    temp['hash'] = '';
    temp['signature'] = '';
    if (authorization != null) {
      temp['authorization'] = authorization!.withoutProofs().toJson();
    }

    // ✅ canonicalize `data` semantically (Map order/whitespace becomes irrelevant)
    final d = temp['data'];
    if (d is Uint8List) {
      final decoded = jsonDecode(utf8.decode(d)); // parse JSON
      temp['data'] = decoded; // store as semantic object
    }

    final canonical = canonicalJsonEncode(temp);
    final encoded = Uint8List.fromList(utf8.encode(canonical));
    final sum = keccak256(encoded);
    return KeyManager.bytesToHex(sum);
  }

  @override
  Transaction get() => this;

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

  static Transaction fromJson(Map<String, dynamic> json) {
    final rawData = json['data'];
    if (rawData == null) {
      throw Exception("transaction data cannot be null");
    }

    return Transaction(
      chainID: json['chain_id'],
      from: json['from'],
      to: json['to'],
      method: json['method'],
      data: rawData,
      version: json['version'],
      uuid7: json['uuid7'],
      hash: json['hash'],
      signature: json['signature'],
      authorization: json['authorization'] == null
          ? null
          : AuthorizationEnvelope.fromJson(
              Map<String, dynamic>.from(json['authorization'] as Map),
            ),
    );
  }
}

/// Signs a transaction using a hex-encoded Ed25519 private key
Future<Transaction> signTransaction(
  String privateKeyHex,
  Transaction tx,
) async {
  await tx.validateUnsignedTransaction();

  final privateKeyBytes = KeyManager.hexToBytes(privateKeyHex);
  if (privateKeyBytes.length < 32) {
    throw Exception("private key must be at least 32 bytes (64 hex chars)");
  }

  final seed = privateKeyBytes.sublist(0, 32);
  final algorithm = Ed25519();
  final keyPair = await algorithm.newKeyPairFromSeed(seed);

  final txHash = await tx.calculateHash();
  final hashBytes = KeyManager.hexToBytes(txHash);
  final signature = await algorithm.sign(hashBytes, keyPair: keyPair);

  tx.hash = txHash;
  final signatureHex = KeyManager.bytesToHex(signature.bytes);
  if (tx.authorization == null) {
    tx.signature = signatureHex;
  } else {
    final publicKey = await keyPair.extractPublicKey();
    tx.authorization!.userSignatures.add(
      AuthorizationSignature(
        publicKey: KeyManager.bytesToHex(publicKey.bytes),
        signature: signatureHex,
      ),
    );
  }
  return tx;
}
