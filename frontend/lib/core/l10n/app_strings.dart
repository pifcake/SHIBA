class AppStrings {
  final String locale;
  const AppStrings(this.locale);

  bool get _ru => locale == 'ru';

  // ── Drawer navigation ────────────────────────────────────────────────────
  String get searchModels => _ru ? 'Поиск моделей' : 'Search Models';
  String get castings => _ru ? 'Кастинги' : 'Castings';
  String get invitations => _ru ? 'Приглашения' : 'Invitations';
  String get myProfile => _ru ? 'Мой профиль' : 'My Profile';
  String get dashboard => _ru ? 'Панель управления' : 'Dashboard';
  String get settingsNav => _ru ? 'Настройки' : 'Settings';
  String get logout => _ru ? 'Выйти' : 'Logout';

  // ── Settings page ────────────────────────────────────────────────────────
  String get settingsTitle => _ru ? 'Настройки' : 'Settings';
  String get languageSection => _ru ? 'Язык интерфейса' : 'Interface Language';
  String get langRussian => _ru ? 'Русский' : 'Russian';
  String get langEnglish => _ru ? 'Английский' : 'English';
  String get themeSection => _ru ? 'Тема оформления' : 'App Theme';
  String get themeLight => _ru ? 'Светлая' : 'Light';
  String get themeDark => _ru ? 'Тёмная' : 'Dark';
  String get themeSystem => _ru ? 'Системная' : 'System';
}
