class AccessPolicy {
  static const String allow = "ALLOW_ACCESS_MODE";
  static const String deny = "DENY_ACCESS_MODE";

  final Map<String, bool> users;
  final String mode;

  AccessPolicy({required this.users, required this.mode})
    : assert(
        mode == allow || mode == deny,
        'mode must be ALLOW_ACCESS_MODE or DENY_ACCESS_MODE',
      ) {
    if (mode != allow && mode != deny) {
      throw ArgumentError('mode must be ALLOW_ACCESS_MODE or DENY_ACCESS_MODE');
    }
  }

  Map<String, dynamic> toJson() => {'users': users, 'mode': mode};

  factory AccessPolicy.fromJson(Map<String, dynamic> json) {
    final mode = json['mode'] as String?;

    if (mode != allow && mode != deny) {
      throw ArgumentError('mode must be ALLOW_ACCESS_MODE or DENY_ACCESS_MODE');
    }

    final rawUsers = json['users'];

    if (rawUsers is! Map) {
      throw ArgumentError('users must be a Map<String, bool>');
    }

    final users = rawUsers.map<String, bool>((key, value) {
      if (value is! bool) {
        throw ArgumentError('users map values must be bool');
      }
      return MapEntry(key.toString(), value);
    });

    return AccessPolicy(users: users, mode: mode!);
  }
}
