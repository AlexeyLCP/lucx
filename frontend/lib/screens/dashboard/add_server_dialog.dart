import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../providers/servers_provider.dart';

final addServerFormProvider =
    StateProvider<Map<String, dynamic>>((ref) => {
          'name': '',
          'host': '',
          'port': 22,
          'username': 'root',
          'auth_method': 'password',
          'credential': '',
        });

class AddServerDialog extends ConsumerWidget {
  const AddServerDialog({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final form = ref.watch(addServerFormProvider);
    return AlertDialog(
      title: const Text('Add Server'),
      content: SingleChildScrollView(
        child: Column(mainAxisSize: MainAxisSize.min, children: [
          TextField(
              decoration: const InputDecoration(labelText: 'Name'),
              onChanged: (v) => _update(ref, 'name', v)),
          const SizedBox(height: 12),
          TextField(
              decoration: const InputDecoration(labelText: 'Host / IP'),
              onChanged: (v) => _update(ref, 'host', v)),
          const SizedBox(height: 12),
          TextField(
              decoration: const InputDecoration(labelText: 'SSH Port'),
              keyboardType: TextInputType.number,
              onChanged: (v) =>
                  _update(ref, 'port', int.tryParse(v) ?? 22)),
          const SizedBox(height: 12),
          TextField(
              decoration: const InputDecoration(labelText: 'Username'),
              onChanged: (v) => _update(ref, 'username', v)),
          const SizedBox(height: 12),
          DropdownButtonFormField<String>(
            value: form['auth_method'],
            decoration: const InputDecoration(labelText: 'Auth Method'),
            items: const [
              DropdownMenuItem(value: 'password', child: Text('Password')),
              DropdownMenuItem(value: 'key', child: Text('SSH Key')),
            ],
            onChanged: (v) => _update(ref, 'auth_method', v ?? 'password'),
          ),
          const SizedBox(height: 12),
          TextField(
            decoration: InputDecoration(
                labelText: form['auth_method'] == 'key'
                    ? 'Private Key'
                    : 'Password'),
            obscureText: form['auth_method'] == 'password',
            maxLines: form['auth_method'] == 'key' ? 5 : 1,
            onChanged: (v) => _update(ref, 'credential', v),
          ),
        ]),
      ),
      actions: [
        TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Cancel')),
        FilledButton(
            onPressed: () {
              ref
                  .read(serversProvider.notifier)
                  .addServer(Map.from(form));
              Navigator.pop(context);
            },
            child: const Text('Add')),
      ],
    );
  }

  void _update(WidgetRef ref, String key, dynamic value) {
    final form = Map<String, dynamic>.from(ref.read(addServerFormProvider));
    form[key] = value;
    ref.read(addServerFormProvider.notifier).state = form;
  }
}
