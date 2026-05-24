import 'package:flutter/material.dart';
import '../models/server.dart';
import 'status_badge.dart';

class ServerCard extends StatelessWidget {
  final Server server;
  final VoidCallback onTap;
  const ServerCard({super.key, required this.server, required this.onTap});

  @override
  Widget build(BuildContext context) {
    return Card(
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(12),
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
            Row(children: [
              Expanded(
                  child: Text(server.name,
                      style: const TextStyle(
                          fontWeight: FontWeight.bold, fontSize: 16))),
              StatusBadge(status: server.status),
            ]),
            const SizedBox(height: 8),
            Text('${server.host}:${server.port}',
                style: const TextStyle(color: Colors.grey, fontSize: 13)),
            const SizedBox(height: 4),
            Text('SSH: ${server.username} (${server.authMethod})',
                style: const TextStyle(color: Colors.grey, fontSize: 12)),
            if (server.os.isNotEmpty)
              Text('OS: ${server.os} ${server.arch}',
                  style: const TextStyle(color: Colors.grey, fontSize: 12)),
          ]),
        ),
      ),
    );
  }
}
