final class SdkError {
  const SdkError({
    required this.error,
    required this.message,
    required this.code,
    this.details = const {},
  });

  factory SdkError.fromJson(Map<String, Object?> json) {
    return SdkError(
      error: json['error']! as String,
      message: json['message']! as String,
      code: json['code']! as String,
      details: Map<String, Object?>.from(json['details'] as Map? ?? const {}),
    );
  }

  final String error;
  final String message;
  final String code;
  final Map<String, Object?> details;
}

final class PaginationResponse {
  const PaginationResponse({
    required this.items,
    required this.limit,
    this.cursor,
    this.nextCursor,
  });

  factory PaginationResponse.fromJson(Map<String, Object?> json) {
    return PaginationResponse(
      items: (json['items']! as List<Object?>)
          .cast<Map<Object?, Object?>>()
          .map((item) => Map<String, Object?>.from(item))
          .toList(),
      limit: json['limit']! as int,
      cursor: json['cursor'] as String?,
      nextCursor: json['next_cursor'] as String?,
    );
  }

  final List<Map<String, Object?>> items;
  final int limit;
  final String? cursor;
  final String? nextCursor;
}

final class IdempotencyRecord {
  const IdempotencyRecord({
    required this.idempotencyKey,
    required this.operation,
    required this.scope,
    required this.requestId,
  });

  factory IdempotencyRecord.fromJson(Map<String, Object?> json) {
    return IdempotencyRecord(
      idempotencyKey: json['idempotency_key']! as String,
      operation: json['operation']! as String,
      scope: json['scope']! as String,
      requestId: json['request_id']! as String,
    );
  }

  final String idempotencyKey;
  final String operation;
  final String scope;
  final String requestId;
}

final class ServiceCatalogEntry {
  const ServiceCatalogEntry({required this.name, required this.env});

  factory ServiceCatalogEntry.fromJson(Map<String, Object?> json) {
    return ServiceCatalogEntry(
      name: json['name']! as String,
      env: json['env']! as String,
    );
  }

  final String name;
  final String env;
}

final class ServiceCatalog {
  const ServiceCatalog({required this.services});

  factory ServiceCatalog.fromJson(Map<String, Object?> json) {
    return ServiceCatalog(
      services: (json['services']! as List<Object?>)
          .cast<Map<Object?, Object?>>()
          .map(
            (item) =>
                ServiceCatalogEntry.fromJson(Map<String, Object?>.from(item)),
          )
          .toList(),
    );
  }

  final List<ServiceCatalogEntry> services;
}

final class ConfiguredServiceEntry {
  const ConfiguredServiceEntry({
    required this.name,
    required this.env,
    required this.url,
  });

  final String name;
  final String env;
  final String url;
}

final class MarketDefinition {
  const MarketDefinition({
    required this.marketId,
    required this.symbol,
    required this.engineId,
    required this.symbolId,
    required this.routeEpoch,
    required this.chain,
    required this.address,
    required this.endpoints,
    required this.status,
    required this.precision,
    required this.tickSize,
    required this.stepSize,
    required this.limits,
    required this.fees,
    required this.capabilities,
    required this.raw,
  });

  factory MarketDefinition.fromJson(Map<String, Object?> json) {
    if (json['schema'] != '2finance.market_definition.v1') {
      throw FormatException('unsupported 2Finance market definition schema');
    }
    return MarketDefinition(
      marketId: json['market_id']! as String,
      symbol: json['symbol']! as String,
      engineId: json['engine_id']! as String,
      symbolId: json['symbol_id']! as int,
      routeEpoch: json['route_epoch']! as int,
      chain: Map<String, Object?>.from(json['chain']! as Map),
      address: json['address']! as String,
      endpoints: Map<String, Object?>.from(json['endpoints']! as Map),
      status: json['status']! as String,
      precision: Map<String, Object?>.from(json['precision']! as Map),
      tickSize: json['tick_size']! as String,
      stepSize: json['step_size']! as String,
      limits: Map<String, Object?>.from(json['limits']! as Map),
      fees: Map<String, Object?>.from(json['fees']! as Map),
      capabilities: Map<String, Object?>.from(json['capabilities']! as Map),
      raw: Map<String, Object?>.from(json),
    );
  }

  final String marketId;
  final String symbol;
  final String engineId;
  final int symbolId;
  final int routeEpoch;
  final Map<String, Object?> chain;
  final String address;
  final Map<String, Object?> endpoints;
  final String status;
  final Map<String, Object?> precision;
  final String tickSize;
  final String stepSize;
  final Map<String, Object?> limits;
  final Map<String, Object?> fees;
  final Map<String, Object?> capabilities;
  final Map<String, Object?> raw;
}

final class MarketDirectory {
  const MarketDirectory({required this.generatedAt, required this.markets});

  factory MarketDirectory.fromJson(Map<String, Object?> json) {
    if (json['schema'] != '2finance.market_directory.v1') {
      throw FormatException('unsupported 2Finance market directory schema');
    }
    return MarketDirectory(
      generatedAt: DateTime.parse(json['generated_at']! as String),
      markets: (json['markets']! as List<Object?>)
          .cast<Map<Object?, Object?>>()
          .map(
            (item) =>
                MarketDefinition.fromJson(Map<String, Object?>.from(item)),
          )
          .toList(growable: false),
    );
  }

  final DateTime generatedAt;
  final List<MarketDefinition> markets;
}

const defaultServiceCatalog = ServiceCatalog(
  services: [
    ServiceCatalogEntry(name: 'auth', env: 'TWO_FINANCE_AUTH_URL'),
    ServiceCatalogEntry(name: 'network', env: 'TWO_FINANCE_NETWORK_URL'),
    ServiceCatalogEntry(name: 'analytics', env: 'TWO_FINANCE_ANALYTICS_URL'),
    ServiceCatalogEntry(
      name: 'orchestrator',
      env: 'TWO_FINANCE_ORCHESTRATOR_URL',
    ),
    ServiceCatalogEntry(name: 'mcp', env: 'TWO_FINANCE_MCP_URL'),
    ServiceCatalogEntry(name: 'planner', env: 'TWO_FINANCE_MCP_URL'),
    ServiceCatalogEntry(
      name: 'tradingcontrol',
      env: 'TWO_FINANCE_TRADING_CONTROL_URL',
    ),
    ServiceCatalogEntry(
      name: 'matchengine',
      env: 'TWO_FINANCE_MATCHENGINE_WS_URL',
    ),
    ServiceCatalogEntry(name: 'keystore', env: 'TWO_FINANCE_KEYSTORE_URL'),
    ServiceCatalogEntry(name: 'hummingbot', env: 'TWO_FINANCE_HUMMINGBOT_URL'),
    ServiceCatalogEntry(name: 'wise', env: 'TWO_FINANCE_WISE_URL'),
    ServiceCatalogEntry(name: 'airwallex', env: 'TWO_FINANCE_AIRWALLEX_URL'),
  ],
);

final class DomainOperation {
  const DomainOperation({
    required this.name,
    required this.method,
    required this.path,
    this.pathParams = const [],
    this.query = const [],
    this.requestSchema,
    this.responseSchema,
    this.notes,
  });

  factory DomainOperation.fromJson(Map<String, Object?> json) {
    return DomainOperation(
      name: json['name']! as String,
      method: json['method']! as String,
      path: json['path']! as String,
      pathParams: (json['path_params'] as List<Object?>? ?? const [])
          .cast<String>(),
      query: (json['query'] as List<Object?>? ?? const []).cast<String>(),
      requestSchema: json['request_schema'] as String?,
      responseSchema: json['response_schema'] as String?,
      notes: json['notes'] as String?,
    );
  }

  final String name;
  final String method;
  final String path;
  final List<String> pathParams;
  final List<String> query;
  final String? requestSchema;
  final String? responseSchema;
  final String? notes;

  ResolvedOperation resolve({
    Map<String, Object?> pathParams = const {},
    Map<String, Object?> queryParameters = const {},
  }) {
    var resolvedPath = path;
    for (final name in this.pathParams) {
      if (!pathParams.containsKey(name)) {
        throw ArgumentError('2finance: missing operation path parameter $name');
      }
      resolvedPath = resolvedPath.replaceAll(
        '{$name}',
        Uri.encodeComponent('${pathParams[name]}'),
      );
    }

    final query = <String, String>{};
    for (final name in this.query) {
      final value = queryParameters[name];
      if (value != null) {
        query[name] = '$value';
      }
    }
    if (query.isNotEmpty) {
      final separator = resolvedPath.contains('?') ? '&' : '?';
      resolvedPath =
          '$resolvedPath$separator${Uri(queryParameters: query).query}';
    }

    return ResolvedOperation(
      method: method.trim().toUpperCase(),
      path: resolvedPath,
    );
  }
}

final class ResolvedOperation {
  const ResolvedOperation({required this.method, required this.path});

  final String method;
  final String path;
}

final class DomainOperationsDomain {
  const DomainOperationsDomain({
    required this.name,
    required this.env,
    required this.operations,
    this.transport,
    this.description,
  });

  factory DomainOperationsDomain.fromJson(Map<String, Object?> json) {
    return DomainOperationsDomain(
      name: json['name']! as String,
      env: json['env']! as String,
      transport: json['transport'] as String?,
      description: json['description'] as String?,
      operations: (json['operations']! as List<Object?>)
          .cast<Map<Object?, Object?>>()
          .map(
            (item) => DomainOperation.fromJson(Map<String, Object?>.from(item)),
          )
          .toList(),
    );
  }

  final String name;
  final String env;
  final String? transport;
  final String? description;
  final List<DomainOperation> operations;
}

final class DomainOperationsCatalog {
  const DomainOperationsCatalog({required this.schema, required this.domains});

  factory DomainOperationsCatalog.fromJson(Map<String, Object?> json) {
    return DomainOperationsCatalog(
      schema: json['schema']! as String,
      domains: (json['domains']! as List<Object?>)
          .cast<Map<Object?, Object?>>()
          .map(
            (item) => DomainOperationsDomain.fromJson(
              Map<String, Object?>.from(item),
            ),
          )
          .toList(),
    );
  }

  final String schema;
  final List<DomainOperationsDomain> domains;

  DomainOperation? operation(String domainName, String operationName) {
    for (final domain in domains) {
      if (_domainKey(domain.name) != _domainKey(domainName)) {
        continue;
      }
      for (final operation in domain.operations) {
        if (operation.name == operationName) {
          return operation;
        }
      }
      return null;
    }
    return null;
  }

  ResolvedOperation resolveOperation(
    String domainName,
    String operationName, {
    Map<String, Object?> pathParams = const {},
    Map<String, Object?> queryParameters = const {},
  }) {
    final found = operation(domainName, operationName);
    if (found == null) {
      throw ArgumentError(
        '2finance: unknown operation $domainName.$operationName',
      );
    }
    return found.resolve(
      pathParams: pathParams,
      queryParameters: queryParameters,
    );
  }
}

String _domainKey(String value) {
  return value.trim().toLowerCase().replaceAll(RegExp(r'[-_\s]+'), '');
}
