import 'package:equatable/equatable.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:shared_preferences/shared_preferences.dart';

import '../../../core/l10n/app_strings.dart';

// ── State ─────────────────────────────────────────────────────────────────────

class SettingsState extends Equatable {
  final ThemeMode themeMode;
  final String locale; // 'ru' | 'en'

  const SettingsState({required this.themeMode, required this.locale});

  AppStrings get strings => AppStrings(locale);

  SettingsState copyWith({ThemeMode? themeMode, String? locale}) =>
      SettingsState(
        themeMode: themeMode ?? this.themeMode,
        locale: locale ?? this.locale,
      );

  @override
  List<Object> get props => [themeMode, locale];
}

// ── Cubit ─────────────────────────────────────────────────────────────────────

class SettingsCubit extends Cubit<SettingsState> {
  final SharedPreferences _prefs;

  static const _keyTheme = 'settings_theme';
  static const _keyLocale = 'settings_locale';

  SettingsCubit(this._prefs) : super(_initState(_prefs));

  static SettingsState _initState(SharedPreferences prefs) {
    final locale = prefs.getString(_keyLocale) ?? 'ru';
    final themeStr = prefs.getString(_keyTheme) ?? 'dark';
    return SettingsState(themeMode: _parseTheme(themeStr), locale: locale);
  }

  static ThemeMode _parseTheme(String s) {
    if (s == 'light') return ThemeMode.light;
    if (s == 'system') return ThemeMode.system;
    return ThemeMode.dark;
  }

  Future<void> setTheme(ThemeMode mode) async {
    final str = mode == ThemeMode.light
        ? 'light'
        : mode == ThemeMode.system
            ? 'system'
            : 'dark';
    await _prefs.setString(_keyTheme, str);
    emit(state.copyWith(themeMode: mode));
  }

  Future<void> setLocale(String locale) async {
    await _prefs.setString(_keyLocale, locale);
    emit(state.copyWith(locale: locale));
  }
}
