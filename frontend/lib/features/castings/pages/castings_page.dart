import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:go_router/go_router.dart';

import '../../../core/di/injection.dart';
import '../../../core/router/app_router.dart';
import '../cubit/castings_cubit.dart';

class CastingsPage extends StatelessWidget {
  const CastingsPage({super.key});

  @override
  Widget build(BuildContext context) {
    return BlocProvider(
      create: (_) => sl<CastingsCubit>()..loadCastings(),
      child: const _CastingsView(),
    );
  }
}

class _CastingsView extends StatefulWidget {
  const _CastingsView();

  @override
  State<_CastingsView> createState() => _CastingsViewState();
}

class _CastingsViewState extends State<_CastingsView> {
  String _status = 'active';
  final _cityCtrl = TextEditingController();

  @override
  void dispose() {
    _cityCtrl.dispose();
    super.dispose();
  }

  void _reload() {
    context.read<CastingsCubit>().loadCastings(
          status: _status,
          city: _cityCtrl.text.trim(),
        );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        leading: Builder(
          builder: (ctx) => IconButton(
            icon: const Icon(Icons.menu),
            onPressed: () => Scaffold.of(ctx).openDrawer(),
          ),
        ),
        title: const Text('Castings'),
        actions: [
          IconButton(
            icon: const Icon(Icons.add),
            tooltip: 'Create Casting',
            onPressed: () => context.go('/castings/create'),
          ),
        ],
      ),
      drawer: const AppDrawer(),
      body: Column(
        children: [
          // Filters
          Padding(
            padding: const EdgeInsets.fromLTRB(12, 12, 12, 0),
            child: Row(
              children: [
                Expanded(
                  child: SegmentedButton<String>(
                    segments: const [
                      ButtonSegment(value: 'active', label: Text('Active')),
                      ButtonSegment(value: 'closed', label: Text('Closed')),
                      ButtonSegment(value: '', label: Text('All')),
                    ],
                    selected: {_status},
                    onSelectionChanged: (s) {
                      setState(() => _status = s.first);
                      _reload();
                    },
                  ),
                ),
              ],
            ),
          ),
          Padding(
            padding: const EdgeInsets.all(12),
            child: Row(
              children: [
                Expanded(
                  child: TextField(
                    controller: _cityCtrl,
                    decoration: const InputDecoration(
                      labelText: 'City filter',
                      prefixIcon: Icon(Icons.location_city),
                      isDense: true,
                    ),
                    onSubmitted: (_) => _reload(),
                  ),
                ),
                const SizedBox(width: 8),
                ElevatedButton(
                  onPressed: _reload,
                  child: const Text('Filter'),
                ),
              ],
            ),
          ),
          const Divider(height: 1),
          // List
          Expanded(
            child: BlocBuilder<CastingsCubit, CastingsState>(
              builder: (context, state) {
                if (state is CastingsLoading) {
                  return const Center(child: CircularProgressIndicator());
                }
                if (state is CastingsError) {
                  return Center(
                    child: Column(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        Text(state.message,
                            style: const TextStyle(color: Colors.red)),
                        const SizedBox(height: 12),
                        ElevatedButton(
                            onPressed: _reload, child: const Text('Retry')),
                      ],
                    ),
                  );
                }
                if (state is CastingsLoaded) {
                  if (state.castings.isEmpty) {
                    return const Center(
                      child: Text('No castings found.',
                          style: TextStyle(color: Colors.grey)),
                    );
                  }
                  return ListView.builder(
                    padding: const EdgeInsets.only(bottom: 80),
                    itemCount: state.castings.length,
                    itemBuilder: (ctx, i) {
                      final c =
                          state.castings[i] as Map<String, dynamic>;
                      return _CastingCard(casting: c);
                    },
                  );
                }
                return const SizedBox();
              },
            ),
          ),
        ],
      ),
    );
  }
}

class _CastingCard extends StatelessWidget {
  final Map<String, dynamic> casting;
  const _CastingCard({required this.casting});

  @override
  Widget build(BuildContext context) {
    final status = casting['status'] as String? ?? '';
    final isActive = status == 'active';

    return Card(
      margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 4),
      child: ListTile(
        leading: CircleAvatar(
          backgroundColor:
              isActive ? Colors.green.shade100 : Colors.grey.shade200,
          child: Icon(
            Icons.work_outline,
            color: isActive ? Colors.green : Colors.grey,
          ),
        ),
        title: Text(
          casting['title'] as String? ?? '',
          style: const TextStyle(fontWeight: FontWeight.w600),
        ),
        subtitle: Text(
          '${casting['city'] ?? ''} • ${status.toUpperCase()}',
          style: const TextStyle(fontSize: 12),
        ),
        trailing: const Icon(Icons.chevron_right),
        onTap: () => context.go('/castings/${casting['id']}'),
      ),
    );
  }
}
