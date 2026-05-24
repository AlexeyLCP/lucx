import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../models/chain.dart';
import '../services/api_client.dart';

final chainsProvider =
    AsyncNotifierProvider<ChainsNotifier, List<Chain>>(() => ChainsNotifier());

class ChainsNotifier extends AsyncNotifier<List<Chain>> {
  @override
  Future<List<Chain>> build() async {
    final client = ref.read(apiClientProvider);
    final data = await client.getChains();
    return data.map((j) => Chain.fromJson(j)).toList();
  }

  Future<void> refresh() async {
    state = const AsyncLoading();
    state = await AsyncValue.guard(() async {
      final client = ref.read(apiClientProvider);
      final data = await client.getChains();
      return data.map((j) => Chain.fromJson(j)).toList();
    });
  }

  Future<Chain> createChain(Map<String, dynamic> data) async {
    final client = ref.read(apiClientProvider);
    final resp = await client.createChain(data);
    await refresh();
    return Chain.fromJson(resp);
  }

  Future<String> applyChain(String id) async {
    final client = ref.read(apiClientProvider);
    final resp = await client.applyChain(id);
    await refresh();
    return resp.toString();
  }
}
