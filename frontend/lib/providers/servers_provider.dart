import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../models/server.dart';
import '../services/api_client.dart';
import 'auth_provider.dart';

final serversProvider =
    AsyncNotifierProvider<ServersNotifier, List<Server>>(() => ServersNotifier());

class ServersNotifier extends AsyncNotifier<List<Server>> {
  @override
  Future<List<Server>> build() async {
    final client = ref.read(apiClientProvider);
    final data = await client.getServers();
    return data.map((j) => Server.fromJson(j)).toList();
  }

  Future<void> refresh() async {
    state = const AsyncLoading();
    state = await AsyncValue.guard(() async {
      final client = ref.read(apiClientProvider);
      final data = await client.getServers();
      return data.map((j) => Server.fromJson(j)).toList();
    });
  }

  Future<void> addServer(Map<String, dynamic> data) async {
    final client = ref.read(apiClientProvider);
    await client.createServer(data);
    await refresh();
  }

  Future<void> removeServer(String id) async {
    final client = ref.read(apiClientProvider);
    await client.deleteServer(id);
    await refresh();
  }
}
