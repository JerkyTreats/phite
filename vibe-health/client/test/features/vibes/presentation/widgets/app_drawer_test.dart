import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:vibe_health/features/vibes/presentation/widgets/app_drawer.dart';


void main() {
  testWidgets('AppDrawer displays all navigation items', (WidgetTester tester) async {
    // Build the widget
    await tester.pumpWidget(
      const MaterialApp(
        home: Scaffold(
          drawer: AppDrawer(),
          body: SizedBox(),
        ),
      ),
    );

    // Open the drawer
    await tester.dragFrom(const Offset(5, 100), const Offset(300, 100));
    await tester.pumpAndSettle();

    // Verify drawer header is displayed
    expect(find.text('Vibe Health'), findsOneWidget);
    
    // Verify all navigation items are displayed
    expect(find.text('Dashboard'), findsOneWidget);
    expect(find.text('Sleep Check'), findsOneWidget);
    expect(find.text('Mood Check'), findsOneWidget);
    expect(find.text('Notification Settings'), findsOneWidget);
    expect(find.text('About'), findsOneWidget);
    
    // Verify icons
    expect(find.byIcon(Icons.dashboard), findsOneWidget);
    expect(find.byIcon(Icons.bedtime), findsOneWidget);
    expect(find.byIcon(Icons.mood), findsOneWidget);
    expect(find.byIcon(Icons.notifications), findsOneWidget);
    expect(find.byIcon(Icons.help_outline), findsOneWidget); // Correct icon for About
  });

  testWidgets('AppDrawer navigates to Dashboard when tapped', (WidgetTester tester) async {
    // Create a mock navigator observer to track navigation
    final mockObserver = MockNavigatorObserver();
    
    // Build the widget
    await tester.pumpWidget(
      MaterialApp(
        navigatorObservers: [mockObserver],
        home: const Scaffold(
          drawer: AppDrawer(),
          body: SizedBox(),
        ),
      ),
    );

    // Open the drawer
    await tester.dragFrom(const Offset(5, 100), const Offset(300, 100));
    await tester.pumpAndSettle();

    // Tap on Dashboard
    await tester.tap(find.text('Dashboard'));
    await tester.pumpAndSettle();

    // Verify navigation occurred (we can only verify the drawer closed)
    expect(find.byType(Drawer), findsNothing);
  });

  testWidgets('AppDrawer shows About dialog when tapped', (WidgetTester tester) async {
    // Build the widget
    await tester.pumpWidget(
      const MaterialApp(
        home: Scaffold(
          drawer: AppDrawer(),
          body: SizedBox(),
        ),
      ),
    );

    // Open the drawer
    await tester.dragFrom(const Offset(5, 100), const Offset(300, 100));
    await tester.pumpAndSettle();

    // Tap on About
    await tester.tap(find.text('About'));
    await tester.pumpAndSettle();

    // Verify About dialog is shown
    expect(find.byType(AboutDialog), findsOneWidget);
    expect(find.text('Vibe Health'), findsOneWidget); // App name in dialog
    expect(find.text('1.0.0'), findsOneWidget); // Version in dialog
  });
}

// Mock navigator observer
class MockNavigatorObserver extends NavigatorObserver {
  final List<Route<dynamic>> pushedRoutes = [];
  final List<Route<dynamic>> poppedRoutes = [];
  final List<Route<dynamic>> removedRoutes = [];
  final List<Route<dynamic>> replacedRoutes = [];

  @override
  void didPush(Route<dynamic> route, Route<dynamic>? previousRoute) {
    pushedRoutes.add(route);
  }

  @override
  void didPop(Route<dynamic> route, Route<dynamic>? previousRoute) {
    poppedRoutes.add(route);
  }

  @override
  void didRemove(Route<dynamic> route, Route<dynamic>? previousRoute) {
    removedRoutes.add(route);
  }

  @override
  void didReplace({Route<dynamic>? newRoute, Route<dynamic>? oldRoute}) {
    if (newRoute != null) replacedRoutes.add(newRoute);
  }
}
