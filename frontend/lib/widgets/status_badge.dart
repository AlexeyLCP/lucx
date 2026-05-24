import 'package:flutter/material.dart';

class StatusBadge extends StatelessWidget {
  final String status;
  const StatusBadge({super.key, required this.status});

  @override
  Widget build(BuildContext context) {
    final color = switch (status) {
      'online' => const Color(0xFF2EA043),
      'offline' => const Color(0xFFDA3633),
      'degraded' => const Color(0xFFBC8C4C),
      'imported' => const Color(0xFF4FC3F7),
      _ => const Color(0xFF8B949E),
    };
    return Row(mainAxisSize: MainAxisSize.min, children: [
      Container(
          width: 8,
          height: 8,
          decoration: BoxDecoration(color: color, shape: BoxShape.circle)),
      const SizedBox(width: 6),
      Text(status,
          style: TextStyle(
              color: color, fontSize: 12, fontWeight: FontWeight.w500)),
    ]);
  }
}
