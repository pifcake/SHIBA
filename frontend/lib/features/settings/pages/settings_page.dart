import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../core/di/injection.dart';
import '../../../core/router/app_router.dart';
import '../cubit/settings_cubit.dart';

class SettingsPage extends StatelessWidget {
  const SettingsPage({super.key});

  @override
  Widget build(BuildContext context) {
    return BlocBuilder<SettingsCubit, SettingsState>(
      bloc: sl<SettingsCubit>(),
      builder: (context, state) {
        final s = state.strings;
        return Scaffold(
          appBar: AppBar(title: Text(s.settingsTitle)),
          drawer: const AppDrawer(),
          body: ListView(
            padding: const EdgeInsets.all(16),
            children: [
              _SectionCard(
                title: s.languageSection,
                children: [
                  RadioListTile<String>(
                    title: Row(children: [
                      const Text('🇷🇺  '),
                      Text(s.langRussian),
                    ]),
                    value: 'ru',
                    groupValue: state.locale,
                    onChanged: (v) => sl<SettingsCubit>().setLocale(v!),
                  ),
                  RadioListTile<String>(
                    title: Row(children: [
                      const Text('🇬🇧  '),
                      Text(s.langEnglish),
                    ]),
                    value: 'en',
                    groupValue: state.locale,
                    onChanged: (v) => sl<SettingsCubit>().setLocale(v!),
                  ),
                ],
              ),
              const SizedBox(height: 16),
              _SectionCard(
                title: s.themeSection,
                children: [
                  RadioListTile<ThemeMode>(
                    secondary: const Icon(Icons.light_mode_outlined),
                    title: Text(s.themeLight),
                    value: ThemeMode.light,
                    groupValue: state.themeMode,
                    onChanged: (v) => sl<SettingsCubit>().setTheme(v!),
                  ),
                  RadioListTile<ThemeMode>(
                    secondary: const Icon(Icons.dark_mode_outlined),
                    title: Text(s.themeDark),
                    value: ThemeMode.dark,
                    groupValue: state.themeMode,
                    onChanged: (v) => sl<SettingsCubit>().setTheme(v!),
                  ),
                  RadioListTile<ThemeMode>(
                    secondary: const Icon(Icons.settings_brightness_outlined),
                    title: Text(s.themeSystem),
                    value: ThemeMode.system,
                    groupValue: state.themeMode,
                    onChanged: (v) => sl<SettingsCubit>().setTheme(v!),
                  ),
                ],
              ),
            ],
          ),
        );
      },
    );
  }
}

class _SectionCard extends StatelessWidget {
  final String title;
  final List<Widget> children;
  const _SectionCard({required this.title, required this.children});

  @override
  Widget build(BuildContext context) {
    return Card(
      child: Padding(
        padding: const EdgeInsets.fromLTRB(16, 16, 16, 8),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(title,
                style: Theme.of(context)
                    .textTheme
                    .titleMedium
                    ?.copyWith(fontWeight: FontWeight.bold)),
            const SizedBox(height: 4),
            ...children,
          ],
        ),
      ),
    );
  }
}
