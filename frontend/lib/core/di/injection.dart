import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:get_it/get_it.dart';
import 'package:shared_preferences/shared_preferences.dart';

import '../api/api_client.dart';
import '../../features/auth/datasource/auth_datasource.dart';
import '../../features/auth/bloc/auth_bloc.dart';
import '../../features/model_profile/cubit/model_profile_cubit.dart';
import '../../features/agency_profile/cubit/agency_profile_cubit.dart';
import '../../features/search/bloc/search_bloc.dart';
import '../../features/castings/cubit/castings_cubit.dart';
import '../../features/invitations/cubit/invitations_cubit.dart';
import '../../features/admin/cubit/admin_cubit.dart';
import '../../features/settings/cubit/settings_cubit.dart';

final GetIt sl = GetIt.instance;

Future<void> configureDependencies() async {
  // Core
  sl.registerLazySingleton<FlutterSecureStorage>(
    () => const FlutterSecureStorage(),
  );

  sl.registerLazySingleton<ApiClient>(
    () => ApiClient(sl<FlutterSecureStorage>()),
  );

  // Settings (singleton — shared across the whole app, loaded from prefs)
  final prefs = await SharedPreferences.getInstance();
  sl.registerLazySingleton<SettingsCubit>(() => SettingsCubit(prefs));

  // Datasources
  sl.registerLazySingleton<AuthDatasource>(
    () => AuthDatasource(sl<ApiClient>()),
  );

  // BLoCs / Cubits (factories so each page gets a fresh instance)
  sl.registerFactory<AuthBloc>(
    () => AuthBloc(sl<AuthDatasource>(), sl<ApiClient>()),
  );

  sl.registerFactory<ModelProfileCubit>(
    () => ModelProfileCubit(sl<ApiClient>()),
  );

  sl.registerFactory<AgencyProfileCubit>(
    () => AgencyProfileCubit(sl<ApiClient>()),
  );

  sl.registerFactory<SearchBloc>(
    () => SearchBloc(sl<ApiClient>()),
  );

  sl.registerFactory<CastingsCubit>(
    () => CastingsCubit(sl<ApiClient>()),
  );

  sl.registerFactory<InvitationsCubit>(
    () => InvitationsCubit(sl<ApiClient>()),
  );

  sl.registerFactory<AdminCubit>(
    () => AdminCubit(sl<ApiClient>()),
  );
}
