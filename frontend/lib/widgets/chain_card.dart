import 'package:flutter/material.dart';
import '../models/chain.dart';

class ChainCard extends StatelessWidget {
  final Chain chain;
  final VoidCallback onTap;
  const ChainCard({super.key, required this.chain, required this.onTap});

  @override
  Widget build(BuildContext context) {
    final color = chain.isActive
        ? const Color(0xFF2EA043)
        : const Color(0xFF8B949E);
    return Card(
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(12),
        child: Padding(
            padding: const EdgeInsets.all(16),
            child: Row(children: [
              Container(
                  width: 4,
                  height: 40,
                  decoration: BoxDecoration(
                      color: color,
                      borderRadius: BorderRadius.circular(2))),
              const SizedBox(width: 12),
              Expanded(
                  child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                    Text(chain.name,
                        style: const TextStyle(fontWeight: FontWeight.bold)),
                    const SizedBox(height: 4),
                    Text(
                      chain.nodes.map((n) => n.role.toUpperCase()).join(' → '),
                      style:
                          const TextStyle(color: Colors.grey, fontSize: 12),
                    ),
                  ])),
              const Icon(Icons.chevron_right, color: Colors.grey),
            ])),
      ),
    );
  }
}
