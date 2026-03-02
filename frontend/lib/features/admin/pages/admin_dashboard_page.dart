import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../core/di/injection.dart';
import '../../../core/router/app_router.dart';
import '../cubit/admin_cubit.dart';

class AdminDashboardPage extends StatelessWidget {
  const AdminDashboardPage({super.key});

  @override
  Widget build(BuildContext context) {
    return BlocProvider(
      create: (_) => sl<AdminCubit>()..loadPendingAgencies(),
      child: const _AdminDashboardView(),
    );
  }
}

class _AdminDashboardView extends StatefulWidget {
  const _AdminDashboardView();

  @override
  State<_AdminDashboardView> createState() => _AdminDashboardViewState();
}

class _AdminDashboardViewState extends State<_AdminDashboardView>
    with SingleTickerProviderStateMixin {
  late TabController _tabController;

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 4, vsync: this);
    _tabController.addListener(() {
      if (_tabController.indexIsChanging) return;
      switch (_tabController.index) {
        case 0:
          context.read<AdminCubit>().loadPendingAgencies();
          break;
        case 1:
          context.read<AdminCubit>().loadPendingPhotos();
          break;
        case 2:
          context.read<AdminCubit>().loadUsers();
          break;
        case 3:
          context.read<AdminCubit>().loadComplaints();
          break;
      }
    });
  }

  @override
  void dispose() {
    _tabController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Admin Dashboard'),
        bottom: TabBar(
          controller: _tabController,
          isScrollable: true,
          tabs: const [
            Tab(text: 'Agencies'),
            Tab(text: 'Photos'),
            Tab(text: 'Users'),
            Tab(text: 'Complaints'),
          ],
        ),
      ),
      drawer: const AppDrawer(),
      body: BlocConsumer<AdminCubit, AdminState>(
        listener: (context, state) {
          if (state is AdminActionSuccess) {
            ScaffoldMessenger.of(context).showSnackBar(
              SnackBar(content: Text(state.message)),
            );
          }
        },
        builder: (context, state) {
          if (state is AdminLoading) {
            return const Center(child: CircularProgressIndicator());
          }
          if (state is AdminError) {
            return Center(
              child: Text(state.message,
                  style: const TextStyle(color: Colors.red)),
            );
          }

          return TabBarView(
            controller: _tabController,
            children: [
              // Agencies tab
              _AgenciesTab(
                agencies:
                    state is AdminAgenciesLoaded ? state.agencies : [],
              ),
              // Photos tab
              _PhotosTab(
                photos: state is AdminPhotosLoaded ? state.photos : [],
              ),
              // Users tab
              _UsersTab(
                users: state is AdminUsersLoaded ? state.users : [],
                total: state is AdminUsersLoaded ? state.total : 0,
              ),
              // Complaints tab
              _ComplaintsTab(
                complaints:
                    state is AdminComplaintsLoaded ? state.complaints : [],
              ),
            ],
          );
        },
      ),
    );
  }
}

// ─── Agencies ───────────────────────────────────────────────────────────────

class _AgenciesTab extends StatelessWidget {
  final List<dynamic> agencies;
  const _AgenciesTab({required this.agencies});

  @override
  Widget build(BuildContext context) {
    if (agencies.isEmpty) {
      return const Center(
        child: Text('No pending agencies.',
            style: TextStyle(color: Colors.grey)),
      );
    }
    return ListView.builder(
      itemCount: agencies.length,
      itemBuilder: (ctx, i) {
        final a = agencies[i] as Map<String, dynamic>;
        final name = a['company_name'] as String? ?? 'Unknown';
        final email = (a['user'] as Map<String, dynamic>?)?['email'] ?? '';
        return Card(
          margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 4),
          child: ListTile(
            leading: const CircleAvatar(child: Icon(Icons.business)),
            title: Text(name,
                style: const TextStyle(fontWeight: FontWeight.w600)),
            subtitle: Text(email as String),
            trailing: Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                IconButton(
                  icon: const Icon(Icons.close, color: Colors.red),
                  tooltip: 'Reject',
                  onPressed: () => context
                      .read<AdminCubit>()
                      .rejectAgency(a['id'] as String),
                ),
                IconButton(
                  icon: const Icon(Icons.check, color: Colors.green),
                  tooltip: 'Approve',
                  onPressed: () => context
                      .read<AdminCubit>()
                      .approveAgency(a['id'] as String),
                ),
              ],
            ),
          ),
        );
      },
    );
  }
}

// ─── Photos ─────────────────────────────────────────────────────────────────

class _PhotosTab extends StatelessWidget {
  final List<dynamic> photos;
  const _PhotosTab({required this.photos});

  @override
  Widget build(BuildContext context) {
    if (photos.isEmpty) {
      return const Center(
        child:
            Text('No photos pending.', style: TextStyle(color: Colors.grey)),
      );
    }
    return ListView.builder(
      itemCount: photos.length,
      itemBuilder: (ctx, i) {
        final p = photos[i] as Map<String, dynamic>;
        final url = p['url'] as String? ?? '';
        return Card(
          margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 4),
          child: Padding(
            padding: const EdgeInsets.all(12),
            child: Row(
              children: [
                // Thumbnail
                ClipRRect(
                  borderRadius: BorderRadius.circular(8),
                  child: url.isNotEmpty
                      ? Image.network(url,
                          width: 80,
                          height: 80,
                          fit: BoxFit.cover,
                          errorBuilder: (_, __, ___) =>
                              const Icon(Icons.broken_image, size: 80))
                      : const Icon(Icons.photo, size: 80),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(p['original_name'] as String? ?? '',
                          style: const TextStyle(
                              fontWeight: FontWeight.w600)),
                      Text(
                          '${((p['size_bytes'] as int? ?? 0) / 1024 / 1024).toStringAsFixed(1)} MB',
                          style: const TextStyle(color: Colors.grey)),
                    ],
                  ),
                ),
                Column(
                  children: [
                    IconButton(
                      icon: const Icon(Icons.thumb_down, color: Colors.red),
                      tooltip: 'Reject',
                      onPressed: () =>
                          _showRejectDialog(context, p['id'] as String),
                    ),
                    IconButton(
                      icon: const Icon(Icons.thumb_up, color: Colors.green),
                      tooltip: 'Approve',
                      onPressed: () => context
                          .read<AdminCubit>()
                          .moderatePhoto(p['id'] as String, 'approved', null),
                    ),
                  ],
                ),
              ],
            ),
          ),
        );
      },
    );
  }

  void _showRejectDialog(BuildContext context, String photoId) {
    final reasonCtrl = TextEditingController();
    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Reject Photo'),
        content: TextField(
          controller: reasonCtrl,
          decoration: const InputDecoration(labelText: 'Reason'),
          maxLines: 2,
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx),
            child: const Text('Cancel'),
          ),
          ElevatedButton(
            style:
                ElevatedButton.styleFrom(backgroundColor: Colors.red),
            onPressed: () {
              context
                  .read<AdminCubit>()
                  .moderatePhoto(photoId, 'rejected', reasonCtrl.text.trim());
              Navigator.pop(ctx);
            },
            child: const Text('Reject'),
          ),
        ],
      ),
    );
  }
}

// ─── Users ──────────────────────────────────────────────────────────────────

class _UsersTab extends StatelessWidget {
  final List<dynamic> users;
  final int total;
  const _UsersTab({required this.users, required this.total});

  @override
  Widget build(BuildContext context) {
    if (users.isEmpty) {
      return const Center(
        child: Text('No users.', style: TextStyle(color: Colors.grey)),
      );
    }
    return Column(
      children: [
        Padding(
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
          child: Row(
            children: [
              Text('$total user(s)',
                  style: const TextStyle(color: Colors.grey)),
            ],
          ),
        ),
        Expanded(
          child: ListView.builder(
            itemCount: users.length,
            itemBuilder: (ctx, i) {
              final u = users[i] as Map<String, dynamic>;
              final email = u['email'] as String? ?? '';
              final role = u['role'] as String? ?? '';
              final status = u['status'] as String? ?? '';
              final photoUrl = u['photo_url'] as String? ?? '';
              final isBlocked = status == 'blocked';

              return Card(
                margin:
                    const EdgeInsets.symmetric(horizontal: 16, vertical: 2),
                child: ListTile(
                  leading: CircleAvatar(
                    backgroundImage: photoUrl.isNotEmpty
                        ? NetworkImage(photoUrl)
                        : null,
                    child: photoUrl.isEmpty
                        ? Text(role.isNotEmpty ? role[0].toUpperCase() : '?')
                        : null,
                  ),
                  title: Text(email),
                  subtitle: Text('$role • $status'),
                  trailing: IconButton(
                    icon: Icon(
                      isBlocked ? Icons.lock_open : Icons.block,
                      color: isBlocked ? Colors.green : Colors.red,
                    ),
                    tooltip: isBlocked ? 'Unblock' : 'Block',
                    onPressed: () => context.read<AdminCubit>().updateUserStatus(
                          u['id'] as String,
                          isBlocked ? 'active' : 'blocked',
                        ),
                  ),
                ),
              );
            },
          ),
        ),
      ],
    );
  }
}

// ─── Complaints ─────────────────────────────────────────────────────────────

class _ComplaintsTab extends StatelessWidget {
  final List<dynamic> complaints;
  const _ComplaintsTab({required this.complaints});

  @override
  Widget build(BuildContext context) {
    if (complaints.isEmpty) {
      return const Center(
        child: Text('No complaints.', style: TextStyle(color: Colors.grey)),
      );
    }
    return ListView.builder(
      itemCount: complaints.length,
      itemBuilder: (ctx, i) {
        final c = complaints[i] as Map<String, dynamic>;
        final status = c['status'] as String? ?? 'open';
        final reason = c['reason_category'] as String? ?? '';
        final desc = c['description'] as String? ?? '';
        final targetType = c['target_type'] as String? ?? '';

        return Card(
          margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 4),
          child: Padding(
            padding: const EdgeInsets.all(12),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  children: [
                    Chip(
                      label: Text(targetType),
                      padding: EdgeInsets.zero,
                    ),
                    const SizedBox(width: 8),
                    Chip(
                      label: Text(reason.replaceAll('_', ' ')),
                      padding: EdgeInsets.zero,
                    ),
                    const Spacer(),
                    Text(status.toUpperCase(),
                        style: TextStyle(
                          color: status == 'open'
                              ? Colors.orange
                              : Colors.grey,
                          fontSize: 12,
                        )),
                  ],
                ),
                if (desc.isNotEmpty) ...[
                  const SizedBox(height: 6),
                  Text(desc),
                ],
                if (status == 'open') ...[
                  const SizedBox(height: 8),
                  Row(
                    children: [
                      OutlinedButton(
                        onPressed: () => context
                            .read<AdminCubit>()
                            .updateComplaint(c['id'] as String, 'dismissed'),
                        child: const Text('Dismiss'),
                      ),
                      const SizedBox(width: 8),
                      ElevatedButton(
                        onPressed: () => context
                            .read<AdminCubit>()
                            .updateComplaint(c['id'] as String, 'resolved'),
                        child: const Text('Resolve'),
                      ),
                    ],
                  ),
                ],
              ],
            ),
          ),
        );
      },
    );
  }
}
