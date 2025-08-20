import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:intl/intl.dart';

import 'package:vibe_health/features/vibes/domain/vibe.dart';
import 'package:vibe_health/features/vibes/domain/vibe_type.dart';
import 'package:vibe_health/features/vibes/presentation/widgets/vibe_list_item.dart';

void main() {
  final testVibe = Vibe(
    id: 'test-id-1',
    userId: 'test-user',
    type: VibeType.mood,
    value: 4,
    note: 'Test note',
    ts: '2025-08-20T12:00:00Z',
  );

  final testVibeNoNote = Vibe(
    id: 'test-id-2',
    userId: 'test-user',
    type: VibeType.sleep,
    value: 3,
    note: null,
    ts: '2025-08-20T08:00:00Z',
  );

  testWidgets('VibeListItem displays correct information with note', (WidgetTester tester) async {
    // Build the widget
    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: VibeListItem(
            vibe: testVibe,
            onTap: () {},
          ),
        ),
      ),
    );

    // Verify icon is displayed
    expect(find.byIcon(Icons.mood), findsOneWidget);

    // Verify type name is displayed
    expect(find.text('Mood'), findsOneWidget);

    // Verify note is displayed
    expect(find.text('Test note'), findsOneWidget);

    // Verify value is displayed
    expect(find.text('4'), findsOneWidget);
    expect(find.text('/ 10'), findsOneWidget);
    
    // Verify description is displayed
    expect(find.text('Feeling okay'), findsOneWidget);

    // Verify timestamp is displayed
    final date = DateTime.parse(testVibe.ts);
    final formattedDate = DateFormat('MMM d, h:mm a').format(date);
    expect(find.text(formattedDate), findsOneWidget);
  });

  testWidgets('VibeListItem displays correctly without note', (WidgetTester tester) async {
    // Build the widget
    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: VibeListItem(
            vibe: testVibeNoNote,
            onTap: () {},
          ),
        ),
      ),
    );

    // Verify icon is displayed
    expect(find.byIcon(Icons.bedtime), findsOneWidget);

    // Verify type name is displayed
    expect(find.text('Sleep'), findsOneWidget);

    // Verify no note is displayed
    expect(find.text('Test note'), findsNothing);

    // Verify value is displayed
    expect(find.text('3'), findsOneWidget);
    expect(find.text('/ 10'), findsOneWidget);
    
    // Verify description is displayed
    expect(find.text('Poor sleep quality'), findsOneWidget);

    // Verify timestamp is displayed
    final date = DateTime.parse(testVibeNoNote.ts);
    final formattedDate = DateFormat('MMM d, h:mm a').format(date);
    expect(find.text(formattedDate), findsOneWidget);
  });

  testWidgets('VibeListItem calls onTap when tapped', (WidgetTester tester) async {
    bool wasTapped = false;
    
    // Build the widget
    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: VibeListItem(
            vibe: testVibe,
            onTap: () {
              wasTapped = true;
            },
          ),
        ),
      ),
    );

    // Tap the widget
    await tester.tap(find.byType(VibeListItem));
    await tester.pump();

    // Verify onTap was called
    expect(wasTapped, true);
  });
}
