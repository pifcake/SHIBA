import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:image_picker/image_picker.dart';

import '../../../core/di/injection.dart';
import '../../../core/router/app_router.dart';
import '../cubit/agency_profile_cubit.dart';

class AgencyProfilePage extends StatelessWidget {
  const AgencyProfilePage({super.key});

  @override
  Widget build(BuildContext context) {
    return BlocProvider(
      create: (_) => sl<AgencyProfileCubit>()..loadProfile(),
      child: const _AgencyProfileView(),
    );
  }
}

class _AgencyProfileView extends StatelessWidget {
  const _AgencyProfileView();

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Agency Profile'),
        actions: [
          IconButton(
            icon: const Icon(Icons.edit),
            onPressed: () => _showEditDialog(context),
          ),
        ],
      ),
      drawer: const AppDrawer(),
      body: BlocBuilder<AgencyProfileCubit, AgencyProfileState>(
        builder: (context, state) {
          if (state is AgencyProfileLoading) {
            return const Center(child: CircularProgressIndicator());
          }
          if (state is AgencyProfileError) {
            return Center(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Text(state.message, style: const TextStyle(color: Colors.red)),
                  const SizedBox(height: 16),
                  ElevatedButton(
                    onPressed: () =>
                        context.read<AgencyProfileCubit>().loadProfile(),
                    child: const Text('Retry'),
                  ),
                ],
              ),
            );
          }
          if (state is AgencyProfileLoaded) {
            return _AgencyContent(profile: state.profile);
          }
          return const SizedBox();
        },
      ),
    );
  }

  void _showEditDialog(BuildContext context) {
    final state = context.read<AgencyProfileCubit>().state;
    if (state is! AgencyProfileLoaded) return;
    showDialog(
      context: context,
      builder: (ctx) => _EditAgencyDialog(
        profile: state.profile,
        cubit: context.read<AgencyProfileCubit>(),
      ),
    );
  }
}

class _AgencyContent extends StatelessWidget {
  final Map<String, dynamic> profile;
  const _AgencyContent({required this.profile});

  @override
  Widget build(BuildContext context) {
    final status = profile['status'] as String? ?? 'pending_approval';
    final isActive = status == 'active';

    return SingleChildScrollView(
      padding: const EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Status banner
          if (!isActive)
            Container(
              width: double.infinity,
              padding: const EdgeInsets.all(12),
              margin: const EdgeInsets.only(bottom: 16),
              decoration: BoxDecoration(
                color: Colors.orange.shade100,
                borderRadius: BorderRadius.circular(8),
                border: Border.all(color: Colors.orange),
              ),
              child: const Row(
                children: [
                  Icon(Icons.hourglass_empty, color: Colors.orange),
                  SizedBox(width: 8),
                  Text(
                    'Your agency is pending approval by admin.',
                    style: TextStyle(color: Colors.orange),
                  ),
                ],
              ),
            ),
          // Logo
          Center(
            child: Column(
              children: [
                GestureDetector(
                  onTap: () => _pickLogo(context),
                  child: CircleAvatar(
                    radius: 60,
                    backgroundImage: profile['logo_url'] != null &&
                            (profile['logo_url'] as String).isNotEmpty
                        ? NetworkImage(profile['logo_url'] as String)
                        : null,
                    child: profile['logo_url'] == null ||
                            (profile['logo_url'] as String).isEmpty
                        ? const Icon(Icons.business, size: 60)
                        : null,
                  ),
                ),
                const SizedBox(height: 8),
                TextButton.icon(
                  icon: const Icon(Icons.upload),
                  label: const Text('Upload Logo'),
                  onPressed: () => _pickLogo(context),
                ),
              ],
            ),
          ),
          const SizedBox(height: 24),
          Text(
            profile['company_name'] as String? ?? 'Company Name',
            style: const TextStyle(fontSize: 24, fontWeight: FontWeight.bold),
          ),
          const SizedBox(height: 8),
          Chip(
            label: Text(status.replaceAll('_', ' ').toUpperCase()),
            backgroundColor: isActive
                ? Colors.green.shade100
                : Colors.orange.shade100,
          ),
          const SizedBox(height: 16),
          if ((profile['description'] as String? ?? '').isNotEmpty) ...[
            const Text('About',
                style: TextStyle(fontWeight: FontWeight.bold, fontSize: 16)),
            const SizedBox(height: 4),
            Text(profile['description'] as String),
            const SizedBox(height: 16),
          ],
          if ((profile['phone'] as String? ?? '').isNotEmpty) ...[
            const Text('Phone',
                style: TextStyle(fontWeight: FontWeight.bold, fontSize: 16)),
            const SizedBox(height: 4),
            Text(profile['phone'] as String),
          ],
        ],
      ),
    );
  }

  Future<void> _pickLogo(BuildContext context) async {
    final picker = ImagePicker();
    final file = await picker.pickImage(source: ImageSource.gallery);
    if (file != null && context.mounted) {
      final bytes = await file.readAsBytes();
      if (!context.mounted) return;
      await context.read<AgencyProfileCubit>().uploadLogo(bytes, file.name);
    }
  }
}

class _EditAgencyDialog extends StatefulWidget {
  final Map<String, dynamic> profile;
  final AgencyProfileCubit cubit;

  const _EditAgencyDialog({required this.profile, required this.cubit});

  @override
  State<_EditAgencyDialog> createState() => _EditAgencyDialogState();
}

class _EditAgencyDialogState extends State<_EditAgencyDialog> {
  late final TextEditingController _name;
  late final TextEditingController _desc;
  late final TextEditingController _phone;

  @override
  void initState() {
    super.initState();
    _name = TextEditingController(
        text: widget.profile['company_name'] as String? ?? '');
    _desc = TextEditingController(
        text: widget.profile['description'] as String? ?? '');
    _phone =
        TextEditingController(text: widget.profile['phone'] as String? ?? '');
  }

  @override
  void dispose() {
    _name.dispose();
    _desc.dispose();
    _phone.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      title: const Text('Edit Agency'),
      content: SingleChildScrollView(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            TextField(
                controller: _name,
                decoration: const InputDecoration(labelText: 'Company Name')),
            const SizedBox(height: 8),
            TextField(
              controller: _desc,
              decoration: const InputDecoration(labelText: 'Description'),
              maxLines: 3,
            ),
            const SizedBox(height: 8),
            TextField(
                controller: _phone,
                decoration: const InputDecoration(labelText: 'Phone')),
          ],
        ),
      ),
      actions: [
        TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Cancel')),
        ElevatedButton(
          onPressed: () {
            widget.cubit.updateProfile({
              'company_name': _name.text,
              'description': _desc.text,
              'phone': _phone.text,
            });
            Navigator.pop(context);
          },
          child: const Text('Save'),
        ),
      ],
    );
  }
}
