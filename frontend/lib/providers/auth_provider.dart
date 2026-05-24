import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../services/api_client.dart';

final apiClientProvider = Provider<ApiClient>((ref) => ApiClient());

final authProvider = StateNotifierProvider<AuthNotifier, bool>((ref) {
  return AuthNotifier(ref.read(apiClientProvider));
});

class AuthNotifier extends StateNotifier<bool> {
  final ApiClient _client;
  AuthNotifier(this._client) : super(false);

  Future<void> login(String password) async {
    await _client.login(password);
    state = true;
  }
}
