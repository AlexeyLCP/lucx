import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../providers/servers_provider.dart';
import '../../providers/auth_provider.dart';
import '../../widgets/server_card.dart';
import 'add_server_dialog.dart';

class DashboardScreen extends ConsumerStatefulWidget {
  const DashboardScreen({super.key});

  @override
  ConsumerState<DashboardScreen> createState() => _DashboardScreenState();
}

class _DashboardScreenState extends ConsumerState<DashboardScreen> {
  final _passwordCtrl = TextEditingController();
  bool _loggedIn = false;

  @override
  Widget build(BuildContext context) {
    final authState = ref.watch(authProvider);
    if (!authState && !_loggedIn) {
      return _buildLogin();
    }
    _loggedIn = true;
    return _buildDashboard();
  }

  Widget _buildLogin() {
    return Scaffold(
      body: Center(
        child: SizedBox(
            width: 320,
            child: Card(
              child: Padding(
                  padding: const EdgeInsets.all(24),
                  child: Column(mainAxisSize: MainAxisSize.min, children: [
                    const Text('LucX',
                        style: TextStyle(
                            fontSize: 24,
                            fontWeight: FontWeight.bold,
                            color: Color(0xFF4FC3F7))),
                    const SizedBox(height: 24),
                    TextField(
                        controller: _passwordCtrl,
                        obscureText: true,
                        decoration: const InputDecoration(
                            labelText: 'Core Password')),
                    const SizedBox(height: 16),
                    SizedBox(
                        width: double.infinity,
                        child: FilledButton(
                            onPressed: () => ref
                                .read(authProvider.notifier)
                                .login(_passwordCtrl.text),
                            child: const Text('Connect to Core'))),
                  ])),
            )),
      ),
    );
  }

  Widget _buildDashboard() {
    final serversAsync = ref.watch(serversProvider);
    return Scaffold(
      appBar: AppBar(title: const Text('LucX'), actions: [
        IconButton(
            icon: const Icon(Icons.refresh),
            onPressed: () =>
                ref.read(serversProvider.notifier).refresh()),
      ]),
      body: serversAsync.when(
        data: (servers) => servers.isEmpty
            ? Center(
                child: Column(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                    const Text('No servers yet',
                        style:
                            TextStyle(fontSize: 18, color: Colors.grey)),
                    const SizedBox(height: 16),
                    FilledButton.icon(
                        icon: const Icon(Icons.add),
                        label: const Text('Add Server'),
                        onPressed: _addServer),
                  ]))
            : ListView.builder(
                padding: const EdgeInsets.all(16),
                itemCount: servers.length,
                itemBuilder: (_, i) => Padding(
                    padding: const EdgeInsets.only(bottom: 8),
                    child:
                        ServerCard(server: servers[i], onTap: () {}))),
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(child: Text('Error: $e')),
      ),
      floatingActionButton: FloatingActionButton(
        onPressed: _addServer,
        child: const Icon(Icons.add),
      ),
    );
  }

  void _addServer() =>
      showDialog(context: context, builder: (_) => const AddServerDialog());
}
