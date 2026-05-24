import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'app/theme.dart';
import 'app/router.dart';

void main() {
  runApp(const ProviderScope(child: LucXApp()));
}

class LucXApp extends StatelessWidget {
  const LucXApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp.router(
      title: 'LucX',
      debugShowCheckedModeBanner: false,
      theme: lucxTheme,
      routerConfig: router,
    );
  }
}
