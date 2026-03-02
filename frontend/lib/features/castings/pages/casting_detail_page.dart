import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../core/di/injection.dart';
import '../../../core/router/app_router.dart';
import '../cubit/castings_cubit.dart';

class CastingDetailPage extends StatelessWidget {
  final String castingId;
  const CastingDetailPage({super.key, required this.castingId});

  @override
  Widget build(BuildContext context) {
    return BlocProvider(
      create: (_) => sl<CastingsCubit>(),
      child: _CastingDetailView(castingId: castingId),
    );
  }
}

class _CastingDetailView extends StatefulWidget {
  final String castingId;
  const _CastingDetailView({required this.castingId});

  @override
  State<_CastingDetailView> createState() => _CastingDetailViewState();
}

class _CastingDetailViewState extends State<_CastingDetailView>
    with SingleTickerProviderStateMixin {
  late TabController _tabController;
  Map<String, dynamic>? _casting;
  List<dynamic> _applications = [];
  bool _loading = true;
  String? _error;

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 2, vsync: this);
    _load();
  }

  @override
  void dispose() {
    _tabController.dispose();
    super.dispose();
  }

  Future<void> _load() async {
    setState(() => _loading = true);
    final cubit = context.read<CastingsCubit>();
    final casting = await cubit.getCasting(widget.castingId);
    final applications = await cubit.getApplications(widget.castingId);
    if (mounted) {
      setState(() {
        _casting = casting;
        _applications = applications;
        _loading = false;
        _error = casting == null ? 'Casting not found' : null;
      });
    }
  }

  void _showApplyDialog() {
    final msgCtrl = TextEditingController();
    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Apply to Casting'),
        content: TextField(
          controller: msgCtrl,
          decoration: const InputDecoration(labelText: 'Message (optional)'),
          maxLines: 3,
        ),
        actions: [
          TextButton(
              onPressed: () => Navigator.pop(ctx),
              child: const Text('Cancel')),
          BlocProvider.value(
            value: context.read<CastingsCubit>(),
            child: Builder(
              builder: (innerCtx) => ElevatedButton(
                onPressed: () async {
                  Navigator.pop(ctx);
                  await innerCtx
                      .read<CastingsCubit>()
                      .applyToCasting(widget.castingId, msgCtrl.text.trim());
                  if (innerCtx.mounted) {
                    ScaffoldMessenger.of(innerCtx).showSnackBar(
                      const SnackBar(
                          content: Text('Application submitted!')),
                    );
                  }
                },
                child: const Text('Submit'),
              ),
            ),
          ),
        ],
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    if (_loading) {
      return Scaffold(
        appBar: AppBar(
          leading: Builder(
            builder: (ctx) => IconButton(
              icon: const Icon(Icons.menu),
              onPressed: () => Scaffold.of(ctx).openDrawer(),
            ),
          ),
          title: const Text('Casting'),
        ),
        drawer: const AppDrawer(),
        body: const Center(child: CircularProgressIndicator()),
      );
    }
    if (_error != null || _casting == null) {
      return Scaffold(
        appBar: AppBar(
          leading: Builder(
            builder: (ctx) => IconButton(
              icon: const Icon(Icons.menu),
              onPressed: () => Scaffold.of(ctx).openDrawer(),
            ),
          ),
          title: const Text('Casting'),
        ),
        drawer: const AppDrawer(),
        body: Center(child: Text(_error ?? 'Unknown error')),
      );
    }

    final c = _casting!;
    final title = c['title'] as String? ?? 'Casting';
    final status = c['status'] as String? ?? '';

    return Scaffold(
      appBar: AppBar(
        leading: Builder(
          builder: (ctx) => IconButton(
            icon: const Icon(Icons.menu),
            onPressed: () => Scaffold.of(ctx).openDrawer(),
          ),
        ),
        title: Text(title),
        bottom: TabBar(
          controller: _tabController,
          tabs: const [
            Tab(text: 'Details'),
            Tab(text: 'Applications'),
          ],
        ),
      ),
      drawer: const AppDrawer(),
      body: TabBarView(
        controller: _tabController,
        children: [
          // Details tab
          SingleChildScrollView(
            padding: const EdgeInsets.all(16),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                // Status chip
                Chip(
                  label: Text(status.toUpperCase()),
                  backgroundColor: status == 'active'
                      ? Colors.green.shade100
                      : Colors.grey.shade200,
                ),
                const SizedBox(height: 16),
                if ((c['city'] as String? ?? '').isNotEmpty) ...[
                  const Text('City',
                      style: TextStyle(
                          fontWeight: FontWeight.bold, fontSize: 16)),
                  const SizedBox(height: 4),
                  Text(c['city'] as String),
                  const SizedBox(height: 12),
                ],
                if (c['casting_date'] != null) ...[
                  const Text('Date',
                      style: TextStyle(
                          fontWeight: FontWeight.bold, fontSize: 16)),
                  const SizedBox(height: 4),
                  Text(c['casting_date'] as String),
                  const SizedBox(height: 12),
                ],
                if ((c['description'] as String? ?? '').isNotEmpty) ...[
                  const Text('Description',
                      style: TextStyle(
                          fontWeight: FontWeight.bold, fontSize: 16)),
                  const SizedBox(height: 4),
                  Text(c['description'] as String),
                ],
              ],
            ),
          ),
          // Applications tab
          _applications.isEmpty
              ? const Center(
                  child: Text('No applications yet.',
                      style: TextStyle(color: Colors.grey)))
              : ListView.builder(
                  itemCount: _applications.length,
                  itemBuilder: (ctx, i) {
                    final app =
                        _applications[i] as Map<String, dynamic>;
                    return _ApplicationTile(
                      application: app,
                      onStatusChange: (appId, st) async {
                        await context
                            .read<CastingsCubit>()
                            .updateApplicationStatus(
                                widget.castingId, appId, st);
                        await _load();
                      },
                    );
                  },
                ),
        ],
      ),
      floatingActionButton: status == 'active'
          ? FloatingActionButton.extended(
              onPressed: _showApplyDialog,
              icon: const Icon(Icons.send),
              label: const Text('Apply'),
            )
          : null,
    );
  }
}

class _ApplicationTile extends StatelessWidget {
  final Map<String, dynamic> application;
  final Future<void> Function(String appId, String status) onStatusChange;

  const _ApplicationTile({
    required this.application,
    required this.onStatusChange,
  });

  @override
  Widget build(BuildContext context) {
    final appId = application['id'] as String? ?? '';
    final status = application['status'] as String? ?? 'pending';
    final message = application['model_message'] as String? ?? '';
    final model = application['model'] as Map<String, dynamic>? ?? {};
    final name =
        '${model['first_name'] ?? ''} ${model['last_name'] ?? ''}'.trim();

    Color statusColor;
    switch (status) {
      case 'accepted':
        statusColor = Colors.green;
        break;
      case 'rejected':
        statusColor = Colors.red;
        break;
      default:
        statusColor = Colors.orange;
    }

    return Card(
      margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 4),
      child: Padding(
        padding: const EdgeInsets.all(12),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Expanded(
                  child: Text(name.isEmpty ? 'Unknown' : name,
                      style: const TextStyle(fontWeight: FontWeight.bold)),
                ),
                Container(
                  padding:
                      const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
                  decoration: BoxDecoration(
                    color: statusColor.withAlpha(30),
                    borderRadius: BorderRadius.circular(12),
                    border: Border.all(color: statusColor),
                  ),
                  child: Text(status.toUpperCase(),
                      style: TextStyle(color: statusColor, fontSize: 11)),
                ),
              ],
            ),
            if (message.isNotEmpty) ...[
              const SizedBox(height: 6),
              Text(message, style: const TextStyle(color: Colors.grey)),
            ],
            if (status == 'pending') ...[
              const SizedBox(height: 8),
              Row(
                children: [
                  OutlinedButton(
                    onPressed: () => onStatusChange(appId, 'rejected'),
                    style: OutlinedButton.styleFrom(
                        foregroundColor: Colors.red),
                    child: const Text('Reject'),
                  ),
                  const SizedBox(width: 8),
                  ElevatedButton(
                    onPressed: () => onStatusChange(appId, 'accepted'),
                    child: const Text('Accept'),
                  ),
                ],
              ),
            ],
          ],
        ),
      ),
    );
  }
}
