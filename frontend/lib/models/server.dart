class Server {
  final String id;
  final String name;
  final String host;
  final int port;
  final String username;
  final String authMethod;
  final String os;
  final String arch;
  final String status;
  final String source;
  final String tags;
  final DateTime? lastSeen;
  final DateTime createdAt;

  Server({
    required this.id,
    required this.name,
    required this.host,
    this.port = 22,
    this.username = 'root',
    this.authMethod = 'password',
    this.os = '',
    this.arch = '',
    this.status = 'unknown',
    this.source = 'fresh',
    this.tags = '[]',
    this.lastSeen,
    DateTime? createdAt,
  }) : createdAt = createdAt ?? DateTime.now();

  factory Server.fromJson(Map<String, dynamic> json) => Server(
        id: json['id'],
        name: json['name'],
        host: json['host'],
        port: json['port'] ?? 22,
        username: json['username'] ?? 'root',
        authMethod: json['auth_method'] ?? 'password',
        os: json['os'] ?? '',
        arch: json['arch'] ?? '',
        status: json['status'] ?? 'unknown',
        source: json['source'] ?? 'fresh',
        tags: json['tags'] ?? '[]',
        lastSeen: json['last_seen'] != null
            ? DateTime.parse(json['last_seen'])
            : null,
        createdAt: json['created_at'] != null
            ? DateTime.parse(json['created_at'])
            : DateTime.now(),
      );

  Map<String, dynamic> toJson() => {
        'id': id,
        'name': name,
        'host': host,
        'port': port,
        'username': username,
        'auth_method': authMethod,
        'os': os,
        'arch': arch,
        'status': status,
        'source': source,
        'tags': tags,
      };

  bool get isOnline => status == 'online';
  bool get isImported => source == 'imported';
}
