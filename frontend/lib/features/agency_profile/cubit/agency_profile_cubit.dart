import 'dart:typed_data';

import 'package:equatable/equatable.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:dio/dio.dart';

import '../../../core/api/api_client.dart';

abstract class AgencyProfileState extends Equatable {
  const AgencyProfileState();
  @override
  List<Object?> get props => [];
}

class AgencyProfileInitial extends AgencyProfileState {}

class AgencyProfileLoading extends AgencyProfileState {}

class AgencyProfileLoaded extends AgencyProfileState {
  final Map<String, dynamic> profile;
  const AgencyProfileLoaded(this.profile);
  @override
  List<Object?> get props => [profile];
}

class AgencyProfileError extends AgencyProfileState {
  final String message;
  const AgencyProfileError(this.message);
  @override
  List<Object?> get props => [message];
}

class AgencyProfileCubit extends Cubit<AgencyProfileState> {
  final ApiClient _api;

  AgencyProfileCubit(this._api) : super(AgencyProfileInitial());

  Future<void> loadProfile() async {
    emit(AgencyProfileLoading());
    try {
      final resp = await _api.get('/agencies/me');
      final profile =
          (resp.data as Map<String, dynamic>)['data'] as Map<String, dynamic>;
      emit(AgencyProfileLoaded(profile));
    } catch (e) {
      emit(AgencyProfileError(e.toString()));
    }
  }

  Future<void> updateProfile(Map<String, dynamic> data) async {
    try {
      final resp = await _api.put('/agencies/me', data: data);
      final profile =
          (resp.data as Map<String, dynamic>)['data'] as Map<String, dynamic>;
      emit(AgencyProfileLoaded(profile));
    } catch (e) {
      emit(AgencyProfileError(e.toString()));
    }
  }

  Future<void> uploadLogo(Uint8List bytes, String filename) async {
    try {
      final formData = FormData.fromMap({
        'logo': MultipartFile.fromBytes(bytes, filename: filename),
      });
      await _api.postMultipart('/agencies/me/logo', formData);
      await loadProfile();
    } catch (e) {
      emit(AgencyProfileError(e.toString()));
    }
  }
}
