import 'package:go_router/go_router.dart';
import '../screens/dashboard/dashboard_screen.dart';
import '../screens/chains/chain_list_screen.dart';
import '../screens/chains/chain_builder_screen.dart';

final router = GoRouter(
  initialLocation: '/',
  routes: [
    GoRoute(
      path: '/',
      builder: (context, state) => const DashboardScreen(),
    ),
    GoRoute(
      path: '/chains',
      builder: (context, state) => const ChainListScreen(),
    ),
    GoRoute(
      path: '/chains/new',
      builder: (context, state) => const ChainBuilderScreen(),
    ),
  ],
);
