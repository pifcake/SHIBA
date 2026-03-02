import 'package:flutter/material.dart';

import '../../../core/api/api_client.dart';
import '../../../core/di/injection.dart';
import '../../../core/router/app_router.dart';

/// Read-only model profile page for agency/admin users.
class ModelViewPage extends StatefulWidget {
  final String modelId;
  const ModelViewPage({super.key, required this.modelId});

  @override
  State<ModelViewPage> createState() => _ModelViewPageState();
}

class _ModelViewPageState extends State<ModelViewPage> {
  final _api = sl<ApiClient>();
  Map<String, dynamic>? _profile;
  List<dynamic> _photos = [];
  bool _loading = true;
  String? _error;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    try {
      final profileResp = await _api.get('/models/${widget.modelId}');
      final profile = (profileResp.data as Map<String, dynamic>)['data']
          as Map<String, dynamic>;

      final photosResp = await _api.get('/models/${widget.modelId}/photos');
      final all = ((photosResp.data as Map<String, dynamic>)['data'] as List?) ?? [];
      // Show only approved photos to agency/admin
      final photos = all
          .where((p) =>
              (p as Map<String, dynamic>)['moderation_status'] == 'approved')
          .toList();

      if (mounted) {
        setState(() {
          _profile = profile;
          _photos = photos;
          _loading = false;
        });
      }
    } catch (e) {
      if (mounted) {
        setState(() {
          _error = e.toString();
          _loading = false;
        });
      }
    }
  }

  String? _coverPhotoUrl() {
    for (final p in _photos) {
      final photo = p as Map<String, dynamic>;
      if (photo['is_cover'] == true && photo['url'] != null) {
        return photo['url'] as String;
      }
    }
    return null;
  }

  @override
  Widget build(BuildContext context) {
    if (_loading) {
      return const Scaffold(
        body: Center(child: CircularProgressIndicator()),
      );
    }
    if (_error != null || _profile == null) {
      return Scaffold(
        appBar: AppBar(title: const Text('Model')),
        drawer: const AppDrawer(),
        body: Center(
          child: Text(_error ?? 'Not found',
              style: const TextStyle(color: Colors.red)),
        ),
      );
    }

    final p = _profile!;
    final name =
        '${p['first_name'] ?? ''} ${p['last_name'] ?? ''}'.trim();

    return Scaffold(
      appBar: AppBar(title: Text(name.isEmpty ? 'Model' : name)),
      drawer: const AppDrawer(),
      body: SingleChildScrollView(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Header
            Container(
              width: double.infinity,
              padding: const EdgeInsets.all(24),
              color: Theme.of(context).colorScheme.primary,
              child: Column(
                children: [
                  CircleAvatar(
                    radius: 50,
                    backgroundImage: _coverPhotoUrl() != null
                        ? NetworkImage(_coverPhotoUrl()!)
                        : null,
                    child: _coverPhotoUrl() == null
                        ? const Icon(Icons.person,
                            size: 50, color: Colors.white)
                        : null,
                  ),
                  const SizedBox(height: 12),
                  Text(
                    name.isEmpty ? 'Unknown' : name,
                    style: const TextStyle(
                        color: Colors.white,
                        fontSize: 22,
                        fontWeight: FontWeight.bold),
                  ),
                  if ((p['city'] as String? ?? '').isNotEmpty)
                    Text(p['city'] as String,
                        style: const TextStyle(color: Colors.white70)),
                ],
              ),
            ),
            // Info
            Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  if ((p['bio'] as String? ?? '').isNotEmpty) ...[
                    _InfoRow('Bio', p['bio'] as String),
                    const Divider(),
                  ],
                  if (p['height_cm'] != null)
                    _InfoRow('Height', '${p['height_cm']} cm'),
                  if (p['weight_kg'] != null)
                    _InfoRow('Weight', '${p['weight_kg']} kg'),
                  if (p['chest_cm'] != null)
                    _InfoRow('Chest', '${p['chest_cm']} cm'),
                  if (p['waist_cm'] != null)
                    _InfoRow('Waist', '${p['waist_cm']} cm'),
                  if (p['hips_cm'] != null)
                    _InfoRow('Hips', '${p['hips_cm']} cm'),
                  if ((p['clothing_size'] as String? ?? '').isNotEmpty)
                    _InfoRow('Clothing size', p['clothing_size'] as String),
                  _InfoRow(
                    'Willing to relocate',
                    (p['willing_to_relocate'] == true) ? 'Yes' : 'No',
                  ),
                  const Divider(),
                  const Text('Photos',
                      style: TextStyle(
                          fontSize: 18, fontWeight: FontWeight.bold)),
                  const SizedBox(height: 8),
                ],
              ),
            ),
            // Photos grid
            if (_photos.isEmpty)
              const Padding(
                padding: EdgeInsets.symmetric(horizontal: 16),
                child: Text('No approved photos.',
                    style: TextStyle(color: Colors.grey)),
              )
            else
              Padding(
                padding: const EdgeInsets.symmetric(horizontal: 16),
                child: GridView.builder(
                  shrinkWrap: true,
                  physics: const NeverScrollableScrollPhysics(),
                  gridDelegate:
                      const SliverGridDelegateWithFixedCrossAxisCount(
                    crossAxisCount: 3,
                    crossAxisSpacing: 4,
                    mainAxisSpacing: 4,
                  ),
                  itemCount: _photos.length,
                  itemBuilder: (ctx, i) {
                    final photo = _photos[i] as Map<String, dynamic>;
                    return Image.network(
                      photo['url'] as String? ?? '',
                      fit: BoxFit.cover,
                      errorBuilder: (_, __, ___) => Container(
                        color: Colors.grey.shade200,
                        child: const Icon(Icons.broken_image),
                      ),
                    );
                  },
                ),
              ),
            const SizedBox(height: 80),
          ],
        ),
      ),
    );
  }
}

class _InfoRow extends StatelessWidget {
  final String label;
  final String value;
  const _InfoRow(this.label, this.value);

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          SizedBox(
            width: 150,
            child: Text(label,
                style: const TextStyle(
                    fontWeight: FontWeight.bold, color: Colors.grey)),
          ),
          Expanded(child: Text(value)),
        ],
      ),
    );
  }
}
