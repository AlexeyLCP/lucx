import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../models/server.dart';
import '../../providers/servers_provider.dart';
import '../../providers/chains_provider.dart';

class ChainBuilderScreen extends ConsumerStatefulWidget {
  const ChainBuilderScreen({super.key});

  @override
  ConsumerState<ChainBuilderScreen> createState() =>
      _ChainBuilderScreenState();
}

class _ChainBuilderScreenState extends ConsumerState<ChainBuilderScreen> {
  final _nameCtrl = TextEditingController();
  final List<String> _selectedServerIds = [];

  @override
  Widget build(BuildContext context) {
    final serversAsync = ref.watch(serversProvider);
    return Scaffold(
      appBar: AppBar(title: const Text('Chain Builder'), actions: [
        TextButton(
          onPressed: _selectedServerIds.length >= 2 ? _applyChain : null,
          child: Text('Apply',
              style: TextStyle(
                  color: _selectedServerIds.length >= 2
                      ? const Color(0xFF2EA043)
                      : Colors.grey)),
        ),
      ]),
      body: serversAsync.when(
        data: (servers) => Column(children: [
          Padding(
              padding: const EdgeInsets.all(16),
              child: TextField(
                controller: _nameCtrl,
                decoration: const InputDecoration(
                    labelText: 'Chain Name',
                    hintText: 'e.g. FI → NL → DE'),
              )),
          const Padding(
              padding: EdgeInsets.symmetric(horizontal: 16),
              child: Text(
                'Select servers in order: Entry → Hop(s) → Exit',
                style: TextStyle(color: Colors.grey, fontSize: 13),
              )),
          const SizedBox(height: 12),
          Expanded(child: _buildCanvas(servers)),
          const SizedBox(height: 80),
        ]),
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(child: Text('Error: $e')),
      ),
    );
  }

  Widget _buildCanvas(List<Server> servers) {
    return ListView.builder(
      padding: const EdgeInsets.symmetric(horizontal: 16),
      itemCount: servers.length,
      itemBuilder: (_, i) {
        final srv = servers[i];
        final isSelected = _selectedServerIds.contains(srv.id);
        final idx = _selectedServerIds.indexOf(srv.id);
        return Card(
          color: isSelected ? const Color(0xFF1A3A1A) : null,
          child: ListTile(
            leading: isSelected
                ? CircleAvatar(
                    backgroundColor: idx == 0
                        ? const Color(0xFF4FC3F7)
                        : idx == _selectedServerIds.length - 1
                            ? const Color(0xFF2EA043)
                            : const Color(0xFFBC8C4C),
                    child: Text('${idx + 1}',
                        style: const TextStyle(fontWeight: FontWeight.bold)),
                  )
                : null,
            title: Text(srv.name),
            subtitle: Text('${srv.host} · ${srv.status}'),
            trailing: isSelected
                ? const Icon(Icons.check_circle, color: Color(0xFF2EA043))
                : null,
            onTap: () => setState(() {
              if (isSelected) {
                _selectedServerIds.remove(srv.id);
              } else {
                _selectedServerIds.add(srv.id);
              }
            }),
          ),
        );
      },
    );
  }

  Future<void> _applyChain() async {
    if (_selectedServerIds.length < 2) {
      ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Select at least 2 servers')));
      return;
    }
    final nodes = _selectedServerIds.asMap().entries.map((e) => {
          'server_id': e.value,
          'backend_type': 'xray',
          'protocol': 'vless',
          'position': e.key,
          'role': e.key == 0
              ? 'entry'
              : e.key == _selectedServerIds.length - 1
                  ? 'exit'
                  : 'hop',
        }).toList();

    final data = {
      'name': _nameCtrl.text.isNotEmpty
          ? _nameCtrl.text
          : 'Chain ${DateTime.now().millisecondsSinceEpoch}',
      'nodes': nodes
    };

    try {
      final chain =
          await ref.read(chainsProvider.notifier).createChain(data);
      await ref.read(chainsProvider.notifier).applyChain(chain.id);
      if (mounted) {
        ScaffoldMessenger.of(context)
            .showSnackBar(const SnackBar(content: Text('Chain applied!')));
        Navigator.pop(context);
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context)
            .showSnackBar(SnackBar(content: Text('Error: $e')));
      }
    }
  }
}
