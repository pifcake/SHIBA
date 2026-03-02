import 'package:equatable/equatable.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../core/api/api_client.dart';

abstract class CastingsState extends Equatable {
  const CastingsState();
  @override
  List<Object?> get props => [];
}

class CastingsInitial extends CastingsState {}

class CastingsLoading extends CastingsState {}

class CastingsLoaded extends CastingsState {
  final List<dynamic> castings;
  final int total;
  const CastingsLoaded({required this.castings, required this.total});
  @override
  List<Object?> get props => [castings, total];
}

class CastingsError extends CastingsState {
  final String message;
  const CastingsError(this.message);
  @override
  List<Object?> get props => [message];
}

class CastingActionSuccess extends CastingsState {
  final String message;
  const CastingActionSuccess(this.message);
  @override
  List<Object?> get props => [message];
}

class CastingsCubit extends Cubit<CastingsState> {
  final ApiClient _api;

  CastingsCubit(this._api) : super(CastingsInitial());

  Future<void> loadCastings({String status = 'active', String city = ''}) async {
    emit(CastingsLoading());
    try {
      final params = <String, dynamic>{'limit': 50, 'offset': 0};
      if (status.isNotEmpty) params['status'] = status;
      if (city.isNotEmpty) params['city'] = city;

      final resp = await _api.get('/castings', queryParameters: params);
      final data =
          (resp.data as Map<String, dynamic>)['data'] as Map<String, dynamic>;
      final castings = (data['data'] as List?) ?? [];
      final total = data['total'] as int? ?? 0;
      emit(CastingsLoaded(castings: castings, total: total));
    } catch (e) {
      emit(CastingsError(e.toString()));
    }
  }

  Future<void> createCasting(Map<String, dynamic> data) async {
    try {
      await _api.post('/castings', data: data);
      emit(const CastingActionSuccess('Casting created'));
      await loadCastings();
    } catch (e) {
      emit(CastingsError(e.toString()));
    }
  }

  Future<void> applyToCasting(String castingId, String message) async {
    try {
      await _api.post('/applications',
          data: {'casting_id': castingId, 'message': message});
      emit(const CastingActionSuccess('Application submitted'));
    } catch (e) {
      emit(CastingsError(e.toString()));
    }
  }

  Future<Map<String, dynamic>?> getCasting(String id) async {
    try {
      final resp = await _api.get('/castings/$id');
      return (resp.data as Map<String, dynamic>)['data'] as Map<String, dynamic>;
    } catch (_) {
      return null;
    }
  }

  Future<List<dynamic>> getApplications(String castingId) async {
    try {
      final resp = await _api.get('/castings/$castingId/applications');
      final data =
          (resp.data as Map<String, dynamic>)['data'] as Map<String, dynamic>;
      return (data['data'] as List?) ?? [];
    } catch (_) {
      return [];
    }
  }

  Future<void> updateApplicationStatus(
      String castingId, String appId, String status) async {
    try {
      await _api.put('/castings/$castingId/applications/$appId',
          data: {'status': status});
      emit(const CastingActionSuccess('Application status updated'));
    } catch (e) {
      emit(CastingsError(e.toString()));
    }
  }
}
