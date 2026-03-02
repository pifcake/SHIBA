import 'package:equatable/equatable.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../core/api/api_client.dart';

abstract class InvitationsState extends Equatable {
  const InvitationsState();
  @override
  List<Object?> get props => [];
}

class InvitationsInitial extends InvitationsState {}

class InvitationsLoading extends InvitationsState {}

class InvitationsLoaded extends InvitationsState {
  final List<dynamic> invitations;
  const InvitationsLoaded(this.invitations);
  @override
  List<Object?> get props => [invitations];
}

class InvitationsError extends InvitationsState {
  final String message;
  const InvitationsError(this.message);
  @override
  List<Object?> get props => [message];
}

class InvitationActionSuccess extends InvitationsState {
  final String message;
  const InvitationActionSuccess(this.message);
  @override
  List<Object?> get props => [message];
}

class InvitationsCubit extends Cubit<InvitationsState> {
  final ApiClient _api;

  InvitationsCubit(this._api) : super(InvitationsInitial());

  Future<void> loadMyInvitations() async {
    emit(InvitationsLoading());
    try {
      final resp = await _api.get('/invitations/me');
      final data =
          (resp.data as Map<String, dynamic>)['data'] as Map<String, dynamic>;
      final invitations = (data['data'] as List?) ?? [];
      emit(InvitationsLoaded(invitations));
    } catch (e) {
      emit(InvitationsError(e.toString()));
    }
  }

  Future<void> loadMyOutgoing() async {
    emit(InvitationsLoading());
    try {
      final resp = await _api.get('/agencies/me/invitations');
      final data =
          (resp.data as Map<String, dynamic>)['data'] as Map<String, dynamic>;
      final invitations = (data['data'] as List?) ?? [];
      emit(InvitationsLoaded(invitations));
    } catch (e) {
      emit(InvitationsError(e.toString()));
    }
  }

  Future<void> respondToInvitation(String id, String status) async {
    try {
      await _api.put('/invitations/$id', data: {'status': status});
      emit(InvitationActionSuccess('Response sent'));
      await loadMyInvitations();
    } catch (e) {
      emit(InvitationsError(e.toString()));
    }
  }

  Future<void> sendInvitation(
      String modelId, String message, String? castingId) async {
    try {
      final body = <String, dynamic>{
        'model_id': modelId,
        'message': message,
      };
      if (castingId != null && castingId.isNotEmpty) {
        body['casting_id'] = castingId;
      }
      await _api.post('/invitations', data: body);
      emit(const InvitationActionSuccess('Invitation sent'));
      await loadMyOutgoing();
    } catch (e) {
      emit(InvitationsError(e.toString()));
    }
  }
}
