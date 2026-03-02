import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:image_picker/image_picker.dart';

import '../../../core/di/injection.dart';
import '../../../core/router/app_router.dart';
import '../cubit/model_profile_cubit.dart';

const _hairColors = [
  '', 'Чёрный', 'Тёмно-коричневый', 'Коричневый', 'Светло-коричневый',
  'Блонд', 'Рыжий', 'Седой', 'Другой',
];
const _hairLengths = ['', 'Короткие', 'Средние', 'Длинные', 'Очень длинные'];
const _hairStructures = ['', 'Прямые', 'Волнистые', 'Кудрявые'];
const _eyeColors = ['', 'Карие', 'Голубые', 'Зелёные', 'Серые', 'Чёрные', 'Смешанный'];

class ModelProfilePage extends StatelessWidget {
  const ModelProfilePage({super.key});

  @override
  Widget build(BuildContext context) {
    return BlocProvider(
      create: (_) => sl<ModelProfileCubit>()..loadProfile(),
      child: const _ModelProfileView(),
    );
  }
}

class _ModelProfileView extends StatelessWidget {
  const _ModelProfileView();

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Мой профиль'),
        actions: [
          IconButton(
            icon: const Icon(Icons.edit),
            onPressed: () => _showEditDialog(context),
          ),
        ],
      ),
      drawer: const AppDrawer(),
      body: BlocBuilder<ModelProfileCubit, ModelProfileState>(
        builder: (context, state) {
          if (state is ModelProfileLoading || state is ModelProfileUpdating) {
            return const Center(child: CircularProgressIndicator());
          }
          if (state is ModelProfileError) {
            return Center(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Text(state.message, style: const TextStyle(color: Colors.red)),
                  const SizedBox(height: 16),
                  ElevatedButton(
                    onPressed: () => context.read<ModelProfileCubit>().loadProfile(),
                    child: const Text('Повторить'),
                  ),
                ],
              ),
            );
          }
          if (state is ModelProfileLoaded) {
            return _ProfileContent(
              profile: state.profile,
              photos: state.photos,
              allCategories: state.allCategories,
            );
          }
          return const SizedBox();
        },
      ),
      floatingActionButton: BlocBuilder<ModelProfileCubit, ModelProfileState>(
        builder: (context, state) {
          if (state is! ModelProfileLoaded) return const SizedBox();
          return FloatingActionButton.extended(
            onPressed: () => _pickAndUploadPhoto(context),
            icon: const Icon(Icons.add_photo_alternate),
            label: const Text('Добавить фото'),
          );
        },
      ),
    );
  }

  Future<void> _pickAndUploadPhoto(BuildContext context) async {
    final picker = ImagePicker();
    final file = await picker.pickImage(source: ImageSource.gallery, imageQuality: 85);
    if (file != null && context.mounted) {
      final mimeType = file.name.endsWith('.png') ? 'image/png' : 'image/jpeg';
      final bytes = await file.readAsBytes();
      if (!context.mounted) return;
      await context.read<ModelProfileCubit>().uploadPhoto(bytes, file.name, mimeType);
    }
  }

  void _showEditDialog(BuildContext context) {
    final state = context.read<ModelProfileCubit>().state;
    if (state is! ModelProfileLoaded) return;
    showDialog(
      context: context,
      builder: (ctx) => _EditProfileDialog(
        profile: state.profile,
        cubit: context.read<ModelProfileCubit>(),
      ),
    );
  }
}

// ─── Profile content ───────────────────────────────────────────────────────

class _ProfileContent extends StatelessWidget {
  final Map<String, dynamic> profile;
  final List<dynamic> photos;
  final List<dynamic> allCategories;

  const _ProfileContent({
    required this.profile,
    required this.photos,
    required this.allCategories,
  });

  @override
  Widget build(BuildContext context) {
    final firstName = profile['first_name'] as String? ?? '';
    final lastName = profile['last_name'] as String? ?? '';
    final city = profile['city'] as String? ?? '';
    final country = profile['country'] as String? ?? '';
    final gender = profile['gender'] as String? ?? '';
    final bio = profile['bio'] as String? ?? '';
    final phone = profile['phone'] as String? ?? '';
    final willing = profile['willing_to_relocate'] as bool? ?? false;
    final profileCategories = (profile['categories'] as List?) ?? [];
    final categoryIds = profileCategories.map((c) => (c as Map)['id']).toSet();

    return SingleChildScrollView(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // ── Header ──
          Container(
            width: double.infinity,
            padding: const EdgeInsets.all(24),
            color: Theme.of(context).colorScheme.primary,
            child: Column(
              children: [
                CircleAvatar(
                  radius: 50,
                  backgroundImage: _getCoverPhoto(),
                  backgroundColor: Colors.white24,
                  child: _getCoverPhoto() == null
                      ? const Icon(Icons.person, size: 50, color: Colors.white)
                      : null,
                ),
                const SizedBox(height: 12),
                Text(
                  '${firstName} ${lastName}'.trim().isEmpty
                      ? 'Ваше имя'
                      : '${firstName} ${lastName}',
                  style: const TextStyle(
                      color: Colors.white,
                      fontSize: 24,
                      fontWeight: FontWeight.bold),
                ),
                if (city.isNotEmpty || country.isNotEmpty)
                  Padding(
                    padding: const EdgeInsets.only(top: 4),
                    child: Text(
                      [city, country].where((s) => s.isNotEmpty).join(', '),
                      style: const TextStyle(color: Colors.white70),
                    ),
                  ),
                if (gender.isNotEmpty) ...[
                  const SizedBox(height: 8),
                  Container(
                    padding:
                        const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
                    decoration: BoxDecoration(
                      color: Colors.white24,
                      borderRadius: BorderRadius.circular(16),
                    ),
                    child: Text(
                      gender == 'male'
                          ? 'Мужчина'
                          : gender == 'female'
                              ? 'Женщина'
                              : gender,
                      style: const TextStyle(color: Colors.white),
                    ),
                  ),
                ],
              ],
            ),
          ),

          // ── Bio / contact ──
          if (bio.isNotEmpty || phone.isNotEmpty)
            Padding(
              padding: const EdgeInsets.fromLTRB(16, 12, 16, 0),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  if (bio.isNotEmpty) _InfoRow('О себе', bio),
                  if (phone.isNotEmpty) _InfoRow('Телефон', phone),
                  _InfoRow('Переезд', willing ? 'Готов(а)' : 'Не готов(а)'),
                ],
              ),
            ),

          const SizedBox(height: 4),

          // ── Параметры тела ──
          _Section(
            title: 'Параметры',
            child: Wrap(
              spacing: 24,
              runSpacing: 16,
              children: [
                if (profile['birth_date'] != null)
                  _ParamChip('Возраст', _calcAge(profile['birth_date'] as String)),
                if (profile['height_cm'] != null)
                  _ParamChip('Рост', '${profile['height_cm']} см'),
                if (profile['weight_kg'] != null)
                  _ParamChip('Вес', '${profile['weight_kg']} кг'),
                if (profile['chest_cm'] != null)
                  _ParamChip('Грудь', '${profile['chest_cm']} см'),
                if (profile['waist_cm'] != null)
                  _ParamChip('Талия', '${profile['waist_cm']} см'),
                if (profile['hips_cm'] != null)
                  _ParamChip('Бёдра', '${profile['hips_cm']} см'),
                if (profile['shoe_size'] != null)
                  _ParamChip('Обувь', '${profile['shoe_size']}'),
                if ((profile['clothing_size_top'] as String? ?? '').isNotEmpty)
                  _ParamChip('Верх', profile['clothing_size_top'] as String),
                if ((profile['clothing_size_bot'] as String? ?? '').isNotEmpty)
                  _ParamChip('Низ', profile['clothing_size_bot'] as String),
              ],
            ),
          ),

          // ── Внешность ──
          _Section(
            title: 'Внешность',
            child: Wrap(
              spacing: 24,
              runSpacing: 16,
              children: [
                if ((profile['hair_color'] as String? ?? '').isNotEmpty)
                  _ParamChip('Цвет волос', profile['hair_color'] as String),
                if ((profile['hair_length'] as String? ?? '').isNotEmpty)
                  _ParamChip('Длина волос', profile['hair_length'] as String),
                if ((profile['hair_structure'] as String? ?? '').isNotEmpty)
                  _ParamChip('Структура', profile['hair_structure'] as String),
                if ((profile['eye_color'] as String? ?? '').isNotEmpty)
                  _ParamChip('Цвет глаз', profile['eye_color'] as String),
              ],
            ),
          ),

          // ── Категории ──
          _Section(
            title: 'Категории работ',
            child: allCategories.isEmpty
                ? const Text('Загрузка...', style: TextStyle(color: Colors.grey))
                : Wrap(
                    spacing: 8,
                    runSpacing: 8,
                    children: allCategories.map((cat) {
                      final catMap = cat as Map<String, dynamic>;
                      final catId = catMap['id'];
                      final isSelected = categoryIds.contains(catId);
                      return FilterChip(
                        label: Text(catMap['name'] as String? ?? ''),
                        selected: isSelected,
                        onSelected: (selected) {
                          final cubit = context.read<ModelProfileCubit>();
                          final id = (catId as num).toInt();
                          if (selected) {
                            cubit.addCategory(id);
                          } else {
                            cubit.removeCategory(id);
                          }
                        },
                      );
                    }).toList(),
                  ),
          ),

          // ── Фотографии ──
          const Padding(
            padding: EdgeInsets.fromLTRB(16, 8, 16, 8),
            child: Text('Фотографии',
                style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold)),
          ),
          if (photos.isEmpty)
            const Padding(
              padding: EdgeInsets.all(16),
              child: Text('Нет фотографий. Добавьте первое фото!',
                  style: TextStyle(color: Colors.grey)),
            )
          else
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 16),
              child: GridView.builder(
                shrinkWrap: true,
                physics: const NeverScrollableScrollPhysics(),
                gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
                  crossAxisCount: 3,
                  crossAxisSpacing: 4,
                  mainAxisSpacing: 4,
                ),
                itemCount: photos.length,
                itemBuilder: (ctx, i) {
                  final photo = photos[i] as Map<String, dynamic>;
                  final photoId = photo['id'] as String;
                  return Stack(
                    fit: StackFit.expand,
                    children: [
                      Image.network(
                        photo['url'] as String? ?? '',
                        fit: BoxFit.cover,
                        errorBuilder: (_, __, ___) => Container(
                          color: Colors.grey.shade200,
                          child: const Icon(Icons.broken_image),
                        ),
                      ),
                      if (photo['moderation_status'] != 'approved')
                        Container(
                          color: Colors.black45,
                          child: Center(
                            child: Text(
                              (photo['moderation_status'] as String?)
                                      ?.toUpperCase() ??
                                  '',
                              style: const TextStyle(
                                  color: Colors.white, fontSize: 10),
                            ),
                          ),
                        ),
                      Positioned(
                        top: 2,
                        right: 2,
                        child: GestureDetector(
                          onTap: () => _confirmDelete(ctx, photoId),
                          child: Container(
                            decoration: const BoxDecoration(
                                color: Colors.black54,
                                shape: BoxShape.circle),
                            padding: const EdgeInsets.all(4),
                            child: const Icon(Icons.close,
                                color: Colors.white, size: 14),
                          ),
                        ),
                      ),
                    ],
                  );
                },
              ),
            ),
          const SizedBox(height: 80),
        ],
      ),
    );
  }

  static String _calcAge(String birthDate) {
    try {
      final dob = DateTime.parse(birthDate);
      final now = DateTime.now();
      int age = now.year - dob.year;
      if (now.month < dob.month ||
          (now.month == dob.month && now.day < dob.day)) age--;
      return '$age лет';
    } catch (_) {
      return birthDate;
    }
  }

  Future<void> _confirmDelete(BuildContext context, String photoId) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Удалить фото?'),
        content: const Text('Это действие нельзя отменить.'),
        actions: [
          TextButton(
              onPressed: () => Navigator.pop(ctx, false),
              child: const Text('Отмена')),
          ElevatedButton(
            style: ElevatedButton.styleFrom(backgroundColor: Colors.red),
            onPressed: () => Navigator.pop(ctx, true),
            child: const Text('Удалить'),
          ),
        ],
      ),
    );
    if (confirmed == true && context.mounted) {
      context.read<ModelProfileCubit>().deletePhoto(photoId);
    }
  }

  ImageProvider? _getCoverPhoto() {
    for (final p in photos) {
      final photo = p as Map<String, dynamic>;
      if (photo['is_cover'] == true && photo['url'] != null) {
        return NetworkImage(photo['url'] as String);
      }
    }
    return null;
  }
}

// ─── Reusable small widgets ────────────────────────────────────────────────

class _Section extends StatelessWidget {
  final String title;
  final Widget child;

  const _Section({required this.title, required this.child});

  @override
  Widget build(BuildContext context) {
    return Theme(
      data: Theme.of(context).copyWith(dividerColor: Colors.transparent),
      child: ExpansionTile(
        initiallyExpanded: true,
        tilePadding: const EdgeInsets.symmetric(horizontal: 16),
        title: Text(title,
            style:
                const TextStyle(fontWeight: FontWeight.bold, fontSize: 16)),
        children: [
          Padding(
            padding: const EdgeInsets.fromLTRB(16, 0, 16, 16),
            child: child,
          ),
        ],
      ),
    );
  }
}

class _ParamChip extends StatelessWidget {
  final String label;
  final String value;
  const _ParamChip(this.label, this.value);

  @override
  Widget build(BuildContext context) {
    return Column(
      mainAxisSize: MainAxisSize.min,
      crossAxisAlignment: CrossAxisAlignment.center,
      children: [
        Text(value,
            style:
                const TextStyle(fontSize: 18, fontWeight: FontWeight.bold)),
        Text(label, style: const TextStyle(fontSize: 11, color: Colors.grey)),
      ],
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
            width: 140,
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

// ─── Edit dialog ───────────────────────────────────────────────────────────

class _EditProfileDialog extends StatefulWidget {
  final Map<String, dynamic> profile;
  final ModelProfileCubit cubit;

  const _EditProfileDialog({required this.profile, required this.cubit});

  @override
  State<_EditProfileDialog> createState() => _EditProfileDialogState();
}

class _EditProfileDialogState extends State<_EditProfileDialog> {
  final _formKey = GlobalKey<FormState>();

  // Basic
  late final TextEditingController _firstName;
  late final TextEditingController _lastName;
  late final TextEditingController _birthDate;
  late final TextEditingController _city;
  late final TextEditingController _country;
  late final TextEditingController _phone;
  late final TextEditingController _bio;
  bool _willing = false;
  String _gender = '';

  // Physical
  late final TextEditingController _heightCm;
  late final TextEditingController _weightKg;
  late final TextEditingController _chestCm;
  late final TextEditingController _waistCm;
  late final TextEditingController _hipsCm;
  late final TextEditingController _shoeSize;
  late final TextEditingController _clothingSizeTop;
  late final TextEditingController _clothingSizeBot;

  // Appearance
  String _hairColor = '';
  String _hairLength = '';
  String _hairStructure = '';
  String _eyeColor = '';

  @override
  void initState() {
    super.initState();
    final p = widget.profile;
    _firstName =
        TextEditingController(text: p['first_name'] as String? ?? '');
    _lastName =
        TextEditingController(text: p['last_name'] as String? ?? '');
    _birthDate =
        TextEditingController(text: p['birth_date'] as String? ?? '');
    _city = TextEditingController(text: p['city'] as String? ?? '');
    _country = TextEditingController(text: p['country'] as String? ?? '');
    _phone = TextEditingController(text: p['phone'] as String? ?? '');
    _bio = TextEditingController(text: p['bio'] as String? ?? '');
    _willing = p['willing_to_relocate'] as bool? ?? false;
    _gender = p['gender'] as String? ?? '';

    _heightCm =
        TextEditingController(text: p['height_cm']?.toString() ?? '');
    _weightKg =
        TextEditingController(text: p['weight_kg']?.toString() ?? '');
    _chestCm =
        TextEditingController(text: p['chest_cm']?.toString() ?? '');
    _waistCm =
        TextEditingController(text: p['waist_cm']?.toString() ?? '');
    _hipsCm = TextEditingController(text: p['hips_cm']?.toString() ?? '');
    _shoeSize =
        TextEditingController(text: p['shoe_size']?.toString() ?? '');
    _clothingSizeTop =
        TextEditingController(text: p['clothing_size_top'] as String? ?? '');
    _clothingSizeBot =
        TextEditingController(text: p['clothing_size_bot'] as String? ?? '');

    _hairColor = _safeOption(p['hair_color'] as String? ?? '', _hairColors);
    _hairLength =
        _safeOption(p['hair_length'] as String? ?? '', _hairLengths);
    _hairStructure =
        _safeOption(p['hair_structure'] as String? ?? '', _hairStructures);
    _eyeColor = _safeOption(p['eye_color'] as String? ?? '', _eyeColors);
  }

  static String _safeOption(String value, List<String> options) =>
      options.contains(value) ? value : '';

  @override
  void dispose() {
    _firstName.dispose();
    _lastName.dispose();
    _birthDate.dispose();
    _city.dispose();
    _country.dispose();
    _phone.dispose();
    _bio.dispose();
    _heightCm.dispose();
    _weightKg.dispose();
    _chestCm.dispose();
    _waistCm.dispose();
    _hipsCm.dispose();
    _shoeSize.dispose();
    _clothingSizeTop.dispose();
    _clothingSizeBot.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Dialog(
      child: ConstrainedBox(
        constraints: BoxConstraints(
          maxWidth: 560,
          maxHeight: MediaQuery.of(context).size.height * 0.92,
        ),
        child: Column(
          children: [
            // Title bar
            Padding(
              padding: const EdgeInsets.fromLTRB(24, 16, 8, 0),
              child: Row(
                children: [
                  const Text('Редактировать профиль',
                      style: TextStyle(
                          fontSize: 20, fontWeight: FontWeight.bold)),
                  const Spacer(),
                  IconButton(
                    icon: const Icon(Icons.close),
                    onPressed: () => Navigator.pop(context),
                  ),
                ],
              ),
            ),
            const Divider(height: 1),
            // Form body
            Expanded(
              child: SingleChildScrollView(
                padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 8),
                child: Form(
                  key: _formKey,
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      // ── Основные данные ──
                      _FormSection('Основные данные'),
                      Row(children: [
                        Expanded(
                          child: TextFormField(
                            controller: _firstName,
                            decoration:
                                const InputDecoration(labelText: 'Имя'),
                          ),
                        ),
                        const SizedBox(width: 12),
                        Expanded(
                          child: TextFormField(
                            controller: _lastName,
                            decoration:
                                const InputDecoration(labelText: 'Фамилия'),
                          ),
                        ),
                      ]),
                      const SizedBox(height: 12),
                      TextFormField(
                        controller: _birthDate,
                        decoration: const InputDecoration(
                          labelText: 'Дата рождения',
                          hintText: 'ГГГГ-ММ-ДД',
                        ),
                      ),
                      const SizedBox(height: 12),
                      // Gender
                      const Text('Пол',
                          style:
                              TextStyle(color: Colors.grey, fontSize: 12)),
                      const SizedBox(height: 6),
                      SegmentedButton<String>(
                        segments: const [
                          ButtonSegment(
                              value: 'male', label: Text('Мужской')),
                          ButtonSegment(
                              value: 'female', label: Text('Женский')),
                        ],
                        selected:
                            _gender.isEmpty ? <String>{} : {_gender},
                        emptySelectionAllowed: true,
                        onSelectionChanged: (s) => setState(
                            () => _gender = s.isEmpty ? '' : s.first),
                      ),
                      const SizedBox(height: 12),
                      Row(children: [
                        Expanded(
                          child: TextFormField(
                            controller: _city,
                            decoration:
                                const InputDecoration(labelText: 'Город'),
                          ),
                        ),
                        const SizedBox(width: 12),
                        Expanded(
                          child: TextFormField(
                            controller: _country,
                            decoration:
                                const InputDecoration(labelText: 'Страна'),
                          ),
                        ),
                      ]),
                      const SizedBox(height: 12),
                      TextFormField(
                        controller: _phone,
                        decoration:
                            const InputDecoration(labelText: 'Телефон'),
                      ),
                      const SizedBox(height: 12),
                      TextFormField(
                        controller: _bio,
                        decoration:
                            const InputDecoration(labelText: 'О себе'),
                        maxLines: 3,
                      ),
                      const SizedBox(height: 4),
                      SwitchListTile(
                        contentPadding: EdgeInsets.zero,
                        title: const Text('Готов(а) к переезду'),
                        value: _willing,
                        onChanged: (v) => setState(() => _willing = v),
                      ),

                      // ── Параметры тела ──
                      _FormSection('Параметры тела'),
                      Row(children: [
                        Expanded(
                            child: _NumField(
                                controller: _heightCm, label: 'Рост (см)')),
                        const SizedBox(width: 12),
                        Expanded(
                            child: _NumField(
                                controller: _weightKg, label: 'Вес (кг)')),
                      ]),
                      const SizedBox(height: 12),
                      Row(children: [
                        Expanded(
                            child: _NumField(
                                controller: _chestCm,
                                label: 'Грудь (см)')),
                        const SizedBox(width: 12),
                        Expanded(
                            child: _NumField(
                                controller: _waistCm,
                                label: 'Талия (см)')),
                        const SizedBox(width: 12),
                        Expanded(
                            child: _NumField(
                                controller: _hipsCm,
                                label: 'Бёдра (см)')),
                      ]),
                      const SizedBox(height: 12),
                      Row(children: [
                        Expanded(
                            child: _NumField(
                                controller: _shoeSize,
                                label: 'Обувь',
                                decimal: true)),
                        const SizedBox(width: 12),
                        Expanded(
                          child: TextFormField(
                            controller: _clothingSizeTop,
                            decoration: const InputDecoration(
                                labelText: 'Верх (р-р)'),
                          ),
                        ),
                        const SizedBox(width: 12),
                        Expanded(
                          child: TextFormField(
                            controller: _clothingSizeBot,
                            decoration: const InputDecoration(
                                labelText: 'Низ (р-р)'),
                          ),
                        ),
                      ]),

                      // ── Внешность ──
                      _FormSection('Внешность'),
                      Row(children: [
                        Expanded(
                          child: _DropField(
                            label: 'Цвет волос',
                            value: _hairColor,
                            items: _hairColors,
                            onChanged: (v) =>
                                setState(() => _hairColor = v ?? ''),
                          ),
                        ),
                        const SizedBox(width: 12),
                        Expanded(
                          child: _DropField(
                            label: 'Длина волос',
                            value: _hairLength,
                            items: _hairLengths,
                            onChanged: (v) =>
                                setState(() => _hairLength = v ?? ''),
                          ),
                        ),
                      ]),
                      const SizedBox(height: 12),
                      Row(children: [
                        Expanded(
                          child: _DropField(
                            label: 'Структура волос',
                            value: _hairStructure,
                            items: _hairStructures,
                            onChanged: (v) =>
                                setState(() => _hairStructure = v ?? ''),
                          ),
                        ),
                        const SizedBox(width: 12),
                        Expanded(
                          child: _DropField(
                            label: 'Цвет глаз',
                            value: _eyeColor,
                            items: _eyeColors,
                            onChanged: (v) =>
                                setState(() => _eyeColor = v ?? ''),
                          ),
                        ),
                      ]),
                      const SizedBox(height: 20),
                    ],
                  ),
                ),
              ),
            ),
            const Divider(height: 1),
            // Actions
            Padding(
              padding: const EdgeInsets.fromLTRB(24, 12, 24, 16),
              child: Row(
                mainAxisAlignment: MainAxisAlignment.end,
                children: [
                  TextButton(
                    onPressed: () => Navigator.pop(context),
                    child: const Text('Отмена'),
                  ),
                  const SizedBox(width: 12),
                  ElevatedButton(
                    onPressed: _save,
                    child: const Text('Сохранить'),
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  void _save() {
    final data = <String, dynamic>{
      'first_name': _firstName.text.trim(),
      'last_name': _lastName.text.trim(),
      'city': _city.text.trim(),
      'country': _country.text.trim(),
      'phone': _phone.text.trim(),
      'bio': _bio.text.trim(),
      'willing_to_relocate': _willing,
      'gender': _gender,
      'hair_color': _hairColor,
      'hair_length': _hairLength,
      'hair_structure': _hairStructure,
      'eye_color': _eyeColor,
      'clothing_size_top': _clothingSizeTop.text.trim(),
      'clothing_size_bot': _clothingSizeBot.text.trim(),
    };
    final bd = _birthDate.text.trim();
    if (bd.isNotEmpty) data['birth_date'] = bd;

    final h = int.tryParse(_heightCm.text.trim());
    if (h != null) data['height_cm'] = h;
    final w = int.tryParse(_weightKg.text.trim());
    if (w != null) data['weight_kg'] = w;
    final ch = int.tryParse(_chestCm.text.trim());
    if (ch != null) data['chest_cm'] = ch;
    final wa = int.tryParse(_waistCm.text.trim());
    if (wa != null) data['waist_cm'] = wa;
    final hi = int.tryParse(_hipsCm.text.trim());
    if (hi != null) data['hips_cm'] = hi;
    final sh = double.tryParse(_shoeSize.text.trim());
    if (sh != null) data['shoe_size'] = sh;

    widget.cubit.updateProfile(data);
    Navigator.pop(context);
  }
}

// ─── Form helper widgets ───────────────────────────────────────────────────

class _FormSection extends StatelessWidget {
  final String title;
  const _FormSection(this.title);

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(top: 20, bottom: 10),
      child: Text(title,
          style: const TextStyle(
              fontSize: 13,
              fontWeight: FontWeight.bold,
              color: Colors.grey,
              letterSpacing: 0.5)),
    );
  }
}

class _NumField extends StatelessWidget {
  final TextEditingController controller;
  final String label;
  final bool decimal;

  const _NumField(
      {required this.controller, required this.label, this.decimal = false});

  @override
  Widget build(BuildContext context) {
    return TextFormField(
      controller: controller,
      decoration: InputDecoration(labelText: label),
      keyboardType: decimal
          ? const TextInputType.numberWithOptions(decimal: true)
          : TextInputType.number,
    );
  }
}

class _DropField extends StatelessWidget {
  final String label;
  final String value;
  final List<String> items;
  final ValueChanged<String?> onChanged;

  const _DropField({
    required this.label,
    required this.value,
    required this.items,
    required this.onChanged,
  });

  @override
  Widget build(BuildContext context) {
    return DropdownButtonFormField<String>(
      value: value,
      decoration: InputDecoration(labelText: label),
      isExpanded: true,
      items: items
          .map((item) => DropdownMenuItem<String>(
                value: item,
                child: Text(
                  item.isEmpty ? 'Не указано' : item,
                  overflow: TextOverflow.ellipsis,
                ),
              ))
          .toList(),
      onChanged: (v) => onChanged(v ?? ''),
    );
  }
}
