import 'dart:typed_data';

import 'package:equatable/equatable.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:dio/dio.dart';

import '../../../core/api/api_client.dart';

// States
abstract class ModelProfileState extends Equatable {
  const ModelProfileState();
  @override
  List<Object?> get props => [];
}

class ModelProfileInitial extends ModelProfileState {}

class ModelProfileLoading extends ModelProfileState {}

class ModelProfileLoaded extends ModelProfileState {
  final Map<String, dynamic> profile;
  final List<dynamic> photos;
  final List<dynamic> allCategories;
  const ModelProfileLoaded({
    required this.profile,
    this.photos = const [],
    this.allCategories = const [],
  });
  @override
  List<Object?> get props => [profile, photos, allCategories];
}

class ModelProfileError extends ModelProfileState {
  final String message;
  const ModelProfileError(this.message);
  @override
  List<Object?> get props => [message];
}

class ModelProfileUpdating extends ModelProfileState {
  final Map<String, dynamic> profile;
  const ModelProfileUpdating(this.profile);
  @override
  List<Object?> get props => [profile];
}

// Cubit
class ModelProfileCubit extends Cubit<ModelProfileState> {
  final ApiClient _api;

  ModelProfileCubit(this._api) : super(ModelProfileInitial());

  Future<void> loadProfile() async {
    emit(ModelProfileLoading());
    try {
      final profileResp = await _api.get('/models/me');
      final profile = (profileResp.data as Map<String, dynamic>)['data']
          as Map<String, dynamic>;

      final profileId = profile['id'] as String;
      final results = await Future.wait([
        _api.get('/models/$profileId/photos'),
        _api.get('/ref/categories'),
      ]);

      final photos =
          ((results[0].data as Map<String, dynamic>)['data'] as List?) ?? [];
      final allCategories =
          ((results[1].data as Map<String, dynamic>)['data'] as List?) ?? [];

      emit(ModelProfileLoaded(
          profile: profile, photos: photos, allCategories: allCategories));
    } catch (e) {
      emit(ModelProfileError(e.toString()));
    }
  }

  Future<void> updateProfile(Map<String, dynamic> data) async {
    final currentState = state;
    if (currentState is ModelProfileLoaded) {
      emit(ModelProfileUpdating(currentState.profile));
    }
    try {
      final resp = await _api.put('/models/me', data: data);
      final profile = (resp.data as Map<String, dynamic>)['data']
          as Map<String, dynamic>;
      final photos = currentState is ModelProfileLoaded ? currentState.photos : [];
      final allCategories = currentState is ModelProfileLoaded ? currentState.allCategories : [];
      emit(ModelProfileLoaded(profile: profile, photos: photos, allCategories: allCategories));
    } catch (e) {
      emit(ModelProfileError(e.toString()));
    }
  }

  Future<void> uploadPhoto(Uint8List bytes, String filename, String mimeType) async {
    try {
      final formData = FormData.fromMap({
        'photo': MultipartFile.fromBytes(bytes,
            filename: filename,
            contentType: DioMediaType.parse(mimeType)),
      });
      await _api.postMultipart('/models/me/photos', formData);
      await loadProfile(); // Reload to get new photo
    } catch (e) {
      emit(ModelProfileError(e.toString()));
    }
  }

  Future<void> deletePhoto(String photoId) async {
    try {
      await _api.delete('/models/me/photos/$photoId');
      await loadProfile();
    } catch (e) {
      emit(ModelProfileError(e.toString()));
    }
  }

  Future<void> addCategory(int categoryId) async {
    try {
      await _api.post('/models/me/categories',
          data: {'category_id': categoryId});
      await loadProfile();
    } catch (e) {
      emit(ModelProfileError(e.toString()));
    }
  }

  Future<void> removeCategory(int categoryId) async {
    try {
      await _api.delete('/models/me/categories/$categoryId');
      await loadProfile();
    } catch (e) {
      emit(ModelProfileError(e.toString()));
    }
  }
}
