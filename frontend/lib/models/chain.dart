class Chain {
  final String id;
  final String name;
  final String status;
  final DateTime? appliedAt;
  final DateTime createdAt;
  final List<ChainNode> nodes;

  Chain({
    required this.id,
    required this.name,
    this.status = 'draft',
    this.appliedAt,
    DateTime? createdAt,
    this.nodes = const [],
  }) : createdAt = createdAt ?? DateTime.now();

  factory Chain.fromJson(Map<String, dynamic> json) => Chain(
        id: json['id'],
        name: json['name'],
        status: json['status'] ?? 'draft',
        appliedAt: json['applied_at'] != null
            ? DateTime.parse(json['applied_at'])
            : null,
        createdAt: json['created_at'] != null
            ? DateTime.parse(json['created_at'])
            : DateTime.now(),
        nodes: (json['nodes'] as List<dynamic>?)
                ?.map((n) => ChainNode.fromJson(n))
                .toList() ??
            [],
      );

  Map<String, dynamic> toJson() => {
        'id': id,
        'name': name,
        'status': status,
        'nodes': nodes.map((n) => n.toJson()).toList(),
      };

  bool get isActive => status == 'active';
  bool get isDraft => status == 'draft';
}

class ChainNode {
  final String chainId;
  final String serverId;
  final String backendType;
  final String protocol;
  final int position;
  final String role;

  ChainNode({
    required this.chainId,
    required this.serverId,
    this.backendType = 'xray',
    this.protocol = 'vless',
    required this.position,
    required this.role,
  });

  factory ChainNode.fromJson(Map<String, dynamic> json) => ChainNode(
        chainId: json['chain_id'] ?? '',
        serverId: json['server_id'],
        backendType: json['backend_type'] ?? 'xray',
        protocol: json['protocol'] ?? 'vless',
        position: json['position'],
        role: json['role'],
      );

  Map<String, dynamic> toJson() => {
        'server_id': serverId,
        'backend_type': backendType,
        'protocol': protocol,
        'position': position,
        'role': role,
      };

  bool get isEntry => role == 'entry';
  bool get isExit => role == 'exit';
  bool get isHop => role == 'hop';
}
