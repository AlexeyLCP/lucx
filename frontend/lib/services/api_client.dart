import 'package:dio/dio.dart';

class ApiClient {
  final Dio _dio;
  String? _token;

  ApiClient({String baseUrl = 'http://localhost:8744'})
      : _dio = Dio(BaseOptions(
          baseUrl: baseUrl,
          connectTimeout: const Duration(seconds: 5),
          receiveTimeout: const Duration(seconds: 10),
          headers: {'Content-Type': 'application/json'},
        ));

  void setToken(String token) {
    _token = token;
    _dio.options.headers['Authorization'] = 'Bearer $token';
  }

  Future<Map<String, dynamic>> login(String password) async {
    final resp = await _dio.post('/api/v1/auth/login', data: {'password': password});
    final token = resp.data['token'];
    setToken(token);
    return resp.data;
  }

  Future<List<dynamic>> getServers() async {
    final resp = await _dio.get('/api/v1/servers');
    return resp.data;
  }

  Future<Map<String, dynamic>> createServer(Map<String, dynamic> data) async {
    final resp = await _dio.post('/api/v1/servers', data: data);
    return resp.data;
  }

  Future<Map<String, dynamic>> getServer(String id) async {
    final resp = await _dio.get('/api/v1/servers/$id');
    return resp.data;
  }

  Future<void> deleteServer(String id) async {
    await _dio.delete('/api/v1/servers/$id');
  }

  Future<Map<String, dynamic>> scanServer(String id) async {
    final resp = await _dio.post('/api/v1/servers/$id/scan');
    return resp.data;
  }

  Future<Map<String, dynamic>> installServer(String id) async {
    final resp = await _dio.post('/api/v1/servers/$id/install');
    return resp.data;
  }

  Future<List<dynamic>> getChains() async {
    final resp = await _dio.get('/api/v1/chains');
    return resp.data;
  }

  Future<Map<String, dynamic>> createChain(Map<String, dynamic> data) async {
    final resp = await _dio.post('/api/v1/chains', data: data);
    return resp.data;
  }

  Future<Map<String, dynamic>> applyChain(String id) async {
    final resp = await _dio.post('/api/v1/chains/$id/apply');
    return resp.data;
  }

  Future<String> getChainConfig(String id) async {
    final resp = await _dio.get('/api/v1/chains/$id/config');
    return resp.data['config'] ?? resp.data.toString();
  }
}
