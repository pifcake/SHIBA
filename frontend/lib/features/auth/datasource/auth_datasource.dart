import '../../../core/api/api_client.dart';

class AuthDatasource {
  final ApiClient _api;
  AuthDatasource(this._api);

  Future<Map<String, dynamic>> register(
      String email, String password, String role) async {
    final r = await _api.post('/auth/register',
        data: {'email': email, 'password': password, 'role': role});
    return r.data as Map<String, dynamic>;
  }

  Future<Map<String, dynamic>> login(String email, String password) async {
    final r = await _api
        .post('/auth/login', data: {'email': email, 'password': password});
    return (r.data as Map<String, dynamic>)['data'] as Map<String, dynamic>;
  }

  Future<void> logout() => _api.post('/auth/logout');
}
