import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../providers/chains_provider.dart';
import '../../widgets/chain_card.dart';
import 'chain_builder_screen.dart';

class ChainListScreen extends ConsumerWidget {
  const ChainListScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final chainsAsync = ref.watch(chainsProvider);
    return Scaffold(
      appBar: AppBar(title: const Text('Chains')),
      body: chainsAsync.when(
        data: (chains) => chains.isEmpty
            ? const Center(
                child: Text('No chains yet. Create your first chain!',
                    style: TextStyle(color: Colors.grey)))
            : ListView.builder(
                padding: const EdgeInsets.all(16),
                itemCount: chains.length,
                itemBuilder: (_, i) => Padding(
                    padding: const EdgeInsets.only(bottom: 8),
                    child: ChainCard(
                        chain: chains[i],
                        onTap: () {
                          Navigator.push(
                              context,
                              MaterialPageRoute(
                                  builder: (_) =>
                                      const ChainBuilderScreen()));
                        })),
              ),
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(child: Text('Error: $e')),
      ),
      floatingActionButton: FloatingActionButton(
        onPressed: () => Navigator.push(context,
            MaterialPageRoute(builder: (_) => const ChainBuilderScreen())),
        child: const Icon(Icons.add),
      ),
    );
  }
}
