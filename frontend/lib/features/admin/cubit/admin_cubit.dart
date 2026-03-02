import 'package:equatable/equatable.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../core/api/api_client.dart';

// States
abstract class AdminState extends Equatable {
  const AdminState();
  @override
  List<Object?> get props => [];
}

class AdminInitial extends AdminState {}

class AdminLoading extends AdminState {}

class AdminUsersLoaded extends AdminState {
  final List<dynamic> users;
  final int total;
  const AdminUsersLoaded({required this.users, required this.total});
  @override
  List<Object?> get props => [users, total];
}

class AdminAgenciesLoaded extends AdminState {
  final List<dynamic> agencies;
  const AdminAgenciesLoaded(this.agencies);
  @override
  List<Object?> get props => [agencies];
}

class AdminPhotosLoaded extends AdminState {
  final List<dynamic> photos;
  const AdminPhotosLoaded(this.photos);
  @override
  List<Object?> get props => [photos];
}

class AdminComplaintsLoaded extends AdminState {
  final List<dynamic> complaints;
  const AdminComplaintsLoaded(this.complaints);
  @override
  List<Object?> get props => [complaints];
}

class AdminActionSuccess extends AdminState {
  final String message;
  const AdminActionSuccess(this.message);
  @override
  List<Object?> get props => [message];
}

class AdminError extends AdminState {
  final String message;
  const AdminError(this.message);
  @override
  List<Object?> get props => [message];
}

class AdminCubit extends Cubit<AdminState> {
  final ApiClient _api;

  AdminCubit(this._api) : super(AdminInitial());

  Future<void> loadUsers({String? role, String? status, int page = 1}) async {
    emit(AdminLoading());
    try {
      final params = <String, dynamic>{'page': page};
      if (role != null && role.isNotEmpty) params['role'] = role;
      if (status != null && status.isNotEmpty) params['status'] = status;

      final resp = await _api.get('/admin/users', queryParameters: params);
      final data =
          (resp.data as Map<String, dynamic>)['data'] as Map<String, dynamic>;
      final users = (data['data'] as List?) ?? [];
      final total = data['total'] as int? ?? 0;
      emit(AdminUsersLoaded(users: users, total: total));
    } catch (e) {
      emit(AdminError(e.toString()));
    }
  }

  Future<void> updateUserStatus(String userId, String status) async {
    try {
      await _api.put('/admin/users/$userId/status', data: {'status': status});
      emit(const AdminActionSuccess('User status updated'));
      await loadUsers();
    } catch (e) {
      emit(AdminError(e.toString()));
    }
  }

  Future<void> loadPendingAgencies() async {
    emit(AdminLoading());
    try {
      final resp = await _api.get('/admin/agencies/pending');
      final data =
          (resp.data as Map<String, dynamic>)['data'] as Map<String, dynamic>;
      final agencies = (data['data'] as List?) ?? [];
      emit(AdminAgenciesLoaded(agencies));
    } catch (e) {
      emit(AdminError(e.toString()));
    }
  }

  Future<void> approveAgency(String agencyId) async {
    try {
      await _api.put('/admin/agencies/$agencyId/approve');
      emit(const AdminActionSuccess('Agency approved'));
      await loadPendingAgencies();
    } catch (e) {
      emit(AdminError(e.toString()));
    }
  }

  Future<void> rejectAgency(String agencyId) async {
    try {
      await _api.put('/admin/agencies/$agencyId/reject');
      emit(const AdminActionSuccess('Agency rejected'));
      await loadPendingAgencies();
    } catch (e) {
      emit(AdminError(e.toString()));
    }
  }

  Future<void> loadPendingPhotos() async {
    emit(AdminLoading());
    try {
      final resp = await _api.get('/admin/photos/pending');
      final data =
          (resp.data as Map<String, dynamic>)['data'] as Map<String, dynamic>;
      final photos = (data['data'] as List?) ?? [];
      emit(AdminPhotosLoaded(photos));
    } catch (e) {
      emit(AdminError(e.toString()));
    }
  }

  Future<void> moderatePhoto(
      String photoId, String status, String? rejectionReason) async {
    try {
      final body = <String, dynamic>{'status': status};
      if (rejectionReason != null && rejectionReason.isNotEmpty) {
        body['rejection_reason'] = rejectionReason;
      }
      await _api.put('/admin/photos/$photoId/moderate', data: body);
      emit(const AdminActionSuccess('Photo moderated'));
      await loadPendingPhotos();
    } catch (e) {
      emit(AdminError(e.toString()));
    }
  }

  Future<void> loadComplaints({String? status}) async {
    emit(AdminLoading());
    try {
      final params = <String, dynamic>{};
      if (status != null && status.isNotEmpty) params['status'] = status;
      final resp =
          await _api.get('/admin/complaints', queryParameters: params);
      final data =
          (resp.data as Map<String, dynamic>)['data'] as Map<String, dynamic>;
      final complaints = (data['data'] as List?) ?? [];
      emit(AdminComplaintsLoaded(complaints));
    } catch (e) {
      emit(AdminError(e.toString()));
    }
  }

  Future<void> updateComplaint(String complaintId, String status) async {
    try {
      await _api.put('/admin/complaints/$complaintId', data: {'status': status});
      emit(const AdminActionSuccess('Complaint updated'));
      await loadComplaints();
    } catch (e) {
      emit(AdminError(e.toString()));
    }
  }
}
