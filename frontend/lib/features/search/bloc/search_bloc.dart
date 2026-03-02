import 'package:equatable/equatable.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../core/api/api_client.dart';

// Events
abstract class SearchEvent extends Equatable {
  const SearchEvent();
  @override
  List<Object?> get props => [];
}

class SearchStarted extends SearchEvent {
  final String city;
  final int? minAge;
  final int? maxAge;
  final int? categoryId;
  final bool? willingToRelocate;
  final int limit;
  final int offset;

  const SearchStarted({
    this.city = '',
    this.minAge,
    this.maxAge,
    this.categoryId,
    this.willingToRelocate,
    this.limit = 20,
    this.offset = 0,
  });

  @override
  List<Object?> get props =>
      [city, minAge, maxAge, categoryId, willingToRelocate, limit, offset];
}

class SearchLoadMore extends SearchEvent {
  const SearchLoadMore();
}

// States
abstract class SearchState extends Equatable {
  const SearchState();
  @override
  List<Object?> get props => [];
}

class SearchInitial extends SearchState {}

class SearchLoading extends SearchState {}

class SearchSuccess extends SearchState {
  final List<dynamic> models;
  final int total;
  final bool hasMore;
  const SearchSuccess(
      {required this.models, required this.total, this.hasMore = false});
  @override
  List<Object?> get props => [models, total, hasMore];
}

class SearchFailure extends SearchState {
  final String message;
  const SearchFailure(this.message);
  @override
  List<Object?> get props => [message];
}

// BLoC
class SearchBloc extends Bloc<SearchEvent, SearchState> {
  final ApiClient _api;
  SearchStarted? _lastFilter;

  SearchBloc(this._api) : super(SearchInitial()) {
    on<SearchStarted>(_onSearch);
    on<SearchLoadMore>(_onLoadMore);
  }

  Future<void> _onSearch(SearchStarted event, Emitter<SearchState> emit) async {
    _lastFilter = event;
    emit(SearchLoading());
    try {
      final params = <String, dynamic>{
        'limit': event.limit,
        'offset': event.offset,
      };
      if (event.city.isNotEmpty) params['city'] = event.city;
      if (event.minAge != null) params['min_age'] = event.minAge;
      if (event.maxAge != null) params['max_age'] = event.maxAge;
      if (event.categoryId != null) params['category_id'] = event.categoryId;
      if (event.willingToRelocate != null) {
        params['willing_to_relocate'] = event.willingToRelocate;
      }

      final resp = await _api.get('/models', queryParameters: params);
      final data = (resp.data as Map<String, dynamic>)['data']
          as Map<String, dynamic>;
      final models = (data['data'] as List?) ?? [];
      final total = data['total'] as int? ?? 0;

      emit(SearchSuccess(
        models: models,
        total: total,
        hasMore: models.length + event.offset < total,
      ));
    } catch (e) {
      emit(SearchFailure(e.toString()));
    }
  }

  Future<void> _onLoadMore(
      SearchLoadMore event, Emitter<SearchState> emit) async {
    final currentState = state;
    final lastFilter = _lastFilter;
    if (currentState is! SearchSuccess || lastFilter == null) return;
    if (!currentState.hasMore) return;

    try {
      final newOffset = currentState.models.length;
      final params = <String, dynamic>{
        'limit': lastFilter.limit,
        'offset': newOffset,
      };
      if (lastFilter.city.isNotEmpty) params['city'] = lastFilter.city;
      if (lastFilter.minAge != null) params['min_age'] = lastFilter.minAge;
      if (lastFilter.maxAge != null) params['max_age'] = lastFilter.maxAge;
      if (lastFilter.categoryId != null) {
        params['category_id'] = lastFilter.categoryId;
      }

      final resp = await _api.get('/models', queryParameters: params);
      final data = (resp.data as Map<String, dynamic>)['data']
          as Map<String, dynamic>;
      final newModels = (data['data'] as List?) ?? [];
      final total = data['total'] as int? ?? currentState.total;

      final allModels = [...currentState.models, ...newModels];
      emit(SearchSuccess(
        models: allModels,
        total: total,
        hasMore: allModels.length < total,
      ));
    } catch (_) {}
  }
}
