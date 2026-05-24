import 'package:flutter/material.dart';

final lucxTheme = ThemeData(
  brightness: Brightness.dark,
  colorScheme: ColorScheme.dark(
    primary: const Color(0xFF1F6FEB),
    secondary: const Color(0xFF4FC3F7),
    surface: const Color(0xFF0D1117),
    error: const Color(0xFFDA3633),
  ),
  scaffoldBackgroundColor: const Color(0xFF0D1117),
  cardTheme: CardTheme(
    color: const Color(0xFF161B22),
    shape: RoundedRectangleBorder(
      borderRadius: BorderRadius.circular(12),
      side: const BorderSide(color: Color(0xFF30363D)),
    ),
  ),
  appBarTheme: const AppBarTheme(
    backgroundColor: Color(0xFF161B22),
    elevation: 0,
  ),
  inputDecorationTheme: InputDecorationTheme(
    filled: true,
    fillColor: const Color(0xFF0D1117),
    border: OutlineInputBorder(
      borderRadius: BorderRadius.circular(8),
      borderSide: const BorderSide(color: Color(0xFF30363D)),
    ),
  ),
);
