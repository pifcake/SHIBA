import 'package:equatable/equatable.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../core/api/api_client.dart';
import '../datasource/auth_datasource.dart';

// Events
abstract class AuthEvent extends Equatable {
  const AuthEvent();
  @override
  List<Object?> get props => [];
}

class LoginRequested extends AuthEvent {
  final String email;
  final String password;
  const LoginRequested(this.email, this.password);
  @override
  List<Object?> get props => [email, password];
}

class RegisterRequested extends AuthEvent {
  final String email;
  final String password;
  final String role;
  const RegisterRequested(this.email, this.password, this.role);
  @override
  List<Object?> get props => [email, password, role];
}

class LogoutRequested extends AuthEvent {}

// States
abstract class AuthState extends Equatable {
  const AuthState();
  @override
  List<Object?> get props => [];
}

class AuthInitial extends AuthState {}

class AuthLoading extends AuthState {}

class AuthSuccess extends AuthState {
  final String role;
  final String message;
  const AuthSuccess({required this.role, this.message = ''});
  @override
  List<Object?> get props => [role, message];
}

class AuthRegistered extends AuthState {
  const AuthRegistered();
}

class AuthFailure extends AuthState {
  final String message;
  const AuthFailure(this.message);
  @override
  List<Object?> get props => [message];
}

class AuthLoggedOut extends AuthState {}

// BLoC
class AuthBloc extends Bloc<AuthEvent, AuthState> {
  final AuthDatasource _datasource;
  final ApiClient _api;

  AuthBloc(this._datasource, this._api) : super(AuthInitial()) {
    on<LoginRequested>(_onLogin);
    on<RegisterRequested>(_onRegister);
    on<LogoutRequested>(_onLogout);
  }

  Future<void> _onLogin(LoginRequested event, Emitter<AuthState> emit) async {
    emit(AuthLoading());
    try {
      final data = await _datasource.login(event.email, event.password);
      final accessToken = data['access_token'] as String;
      final refreshToken = data['refresh_token'] as String;
      final role = data['role'] as String;
      await _api.saveTokens(accessToken, refreshToken);
      await _api.saveRole(role);
      emit(AuthSuccess(role: role));
    } catch (e) {
      emit(AuthFailure(_extractError(e)));
    }
  }

  Future<void> _onRegister(RegisterRequested event, Emitter<AuthState> emit) async {
    emit(AuthLoading());
    try {
      await _datasource.register(event.email, event.password, event.role);
      emit(const AuthRegistered());
    } catch (e) {
      emit(AuthFailure(_extractError(e)));
    }
  }

  Future<void> _onLogout(LogoutRequested event, Emitter<AuthState> emit) async {
    try {
      await _datasource.logout();
    } catch (_) {}
    await _api.clearTokens();
    emit(AuthLoggedOut());
  }

  String _extractError(Object e) {
    return e.toString().replaceAll('Exception: ', '');
  }
}
