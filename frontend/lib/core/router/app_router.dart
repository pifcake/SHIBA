import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:go_router/go_router.dart';

import '../api/api_client.dart';
import '../di/injection.dart';
import '../../features/auth/pages/login_page.dart';
import '../../features/auth/pages/register_page.dart';
import '../../features/auth/pages/verify_email_page.dart';
import '../../features/model_profile/pages/model_profile_page.dart';
import '../../features/model_profile/pages/model_view_page.dart';
import '../../features/agency_profile/pages/agency_profile_page.dart';
import '../../features/search/pages/search_page.dart';
import '../../features/castings/pages/castings_page.dart';
import '../../features/castings/pages/casting_detail_page.dart';
import '../../features/castings/pages/create_casting_page.dart';
import '../../features/invitations/pages/invitations_page.dart';
import '../../features/admin/pages/admin_dashboard_page.dart';
import '../../features/settings/pages/settings_page.dart';
import '../../features/settings/cubit/settings_cubit.dart';

class AppRouter {
  static final router = GoRouter(
    initialLocation: '/login',
    routes: [
      GoRoute(path: '/login', builder: (ctx, state) => const LoginPage()),
      GoRoute(path: '/register', builder: (ctx, state) => const RegisterPage()),
      GoRoute(path: '/verify-email', builder: (ctx, state) => const VerifyEmailPage()),

      ShellRoute(
        builder: (ctx, state, child) => child,
        routes: [
          GoRoute(path: '/profile/model', builder: (ctx, state) => const ModelProfilePage()),
          GoRoute(path: '/profile/agency', builder: (ctx, state) => const AgencyProfilePage()),
          GoRoute(path: '/search', builder: (ctx, state) => const SearchPage()),
          GoRoute(path: '/castings', builder: (ctx, state) => const CastingsPage()),
          GoRoute(path: '/castings/create', builder: (ctx, state) => const CreateCastingPage()),
          GoRoute(
            path: '/castings/:id',
            builder: (ctx, state) => CastingDetailPage(castingId: state.pathParameters['id']!),
          ),
          GoRoute(
            path: '/models/:id',
            builder: (ctx, state) => ModelViewPage(modelId: state.pathParameters['id']!),
          ),
          GoRoute(path: '/invitations', builder: (ctx, state) => const InvitationsPage()),
          GoRoute(path: '/admin', builder: (ctx, state) => const AdminDashboardPage()),
          GoRoute(path: '/settings', builder: (ctx, state) => const SettingsPage()),
        ],
      ),
    ],
  );
}

class AppDrawer extends StatelessWidget {
  const AppDrawer({super.key});

  @override
  Widget build(BuildContext context) {
    return Drawer(
      child: BlocBuilder<SettingsCubit, SettingsState>(
        bloc: sl<SettingsCubit>(),
        builder: (context, settings) {
          final s = settings.strings;
          return FutureBuilder<String?>(
            future: sl<ApiClient>().getRole(),
            builder: (context, snapshot) {
              final role = snapshot.data ?? '';
              return ListView(
                padding: EdgeInsets.zero,
                children: [
                  const DrawerHeader(
                    decoration: BoxDecoration(color: Color(0xFF1A1A2E)),
                    child: Text(
                      'SHIBA',
                      style: TextStyle(
                          color: Colors.white,
                          fontSize: 24,
                          fontWeight: FontWeight.bold),
                    ),
                  ),
                  if (role == 'agency' || role == 'admin')
                    ListTile(
                      leading: const Icon(Icons.search),
                      title: Text(s.searchModels),
                      onTap: () {
                        Navigator.pop(context);
                        context.go('/search');
                      },
                    ),
                  ListTile(
                    leading: const Icon(Icons.work),
                    title: Text(s.castings),
                    onTap: () {
                      Navigator.pop(context);
                      context.go('/castings');
                    },
                  ),
                  if (role == 'model' || role == 'agency')
                    ListTile(
                      leading: const Icon(Icons.mail_outline),
                      title: Text(s.invitations),
                      onTap: () {
                        Navigator.pop(context);
                        context.go('/invitations');
                      },
                    ),
                  if (role == 'model')
                    ListTile(
                      leading: const Icon(Icons.person),
                      title: Text(s.myProfile),
                      onTap: () {
                        Navigator.pop(context);
                        context.go('/profile/model');
                      },
                    ),
                  if (role == 'agency')
                    ListTile(
                      leading: const Icon(Icons.business),
                      title: Text(s.myProfile),
                      onTap: () {
                        Navigator.pop(context);
                        context.go('/profile/agency');
                      },
                    ),
                  if (role == 'admin')
                    ListTile(
                      leading: const Icon(Icons.admin_panel_settings),
                      title: Text(s.dashboard),
                      onTap: () {
                        Navigator.pop(context);
                        context.go('/admin');
                      },
                    ),
                  const Divider(),
                  ListTile(
                    leading: const Icon(Icons.settings_outlined),
                    title: Text(s.settingsNav),
                    onTap: () {
                      Navigator.pop(context);
                      context.go('/settings');
                    },
                  ),
                  ListTile(
                    leading: const Icon(Icons.logout),
                    title: Text(s.logout),
                    onTap: () async {
                      await sl<ApiClient>().clearTokens();
                      if (context.mounted) context.go('/login');
                    },
                  ),
                ],
              );
            },
          );
        },
      ),
    );
  }
}
