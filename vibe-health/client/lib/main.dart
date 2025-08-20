import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:app_links/app_links.dart';

import 'core/di/providers.dart';
import 'core/notifications/notification_service.dart';
import 'core/auth/auth_service.dart';
import 'features/vibes/domain/vibe_type.dart';
import 'features/vibes/presentation/screens/vibe_dashboard_screen.dart';
import 'features/vibes/presentation/screens/vibe_entry_screen.dart';
import 'features/vibes/presentation/screens/notification_settings_screen.dart';

Future<void> main() async {
  // Ensure Flutter is initialized
  WidgetsFlutterBinding.ensureInitialized();

  // Initialize shared preferences
  final sharedPreferences = await SharedPreferences.getInstance();

  // Run the app with providers
  runApp(
    ProviderScope(
      overrides: [
        // Override the shared preferences provider with the instance
        sharedPreferencesProvider.overrideWithValue(sharedPreferences),
      ],
      child: const MyApp(),
    ),
  );
}

class MyApp extends ConsumerStatefulWidget {
  const MyApp({super.key});

  @override
  ConsumerState<MyApp> createState() => _MyAppState();
}

class _MyAppState extends ConsumerState<MyApp> with WidgetsBindingObserver {
  StreamSubscription? _deepLinkSubscription;
  bool _initialUriHandled = false;
  late AppLinks _appLinks;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addObserver(this);
    _initializeApp();
  }

  @override
  void dispose() {
    WidgetsBinding.instance.removeObserver(this);
    _deepLinkSubscription?.cancel();
    super.dispose();
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    if (state == AppLifecycleState.resumed) {
      _handleIncomingLinks();
    }
  }

  Future<void> _initializeApp() async {
    // Initialize notification service
    final notificationService = ref.read(notificationServiceProvider);
    await notificationService.initialize();
    
    // Initialize auth service
    final authService = ref.read(authServiceProvider);
    await authService.initialize();
    
    // For development purposes, set a mock token if not authenticated
    if (!authService.isAuthenticated) {
      await authService.setMockToken('dev-token-123');
    }

    // Initialize app links handler
    _appLinks = AppLinks();

    // Handle deep links
    if (!_initialUriHandled) {
      _handleInitialUri();
      _initialUriHandled = true;
    }
    _handleIncomingLinks();
  }

  Future<void> _handleInitialUri() async {
    try {
      final uri = await _appLinks.getInitialLink();
      if (uri != null) {
        _handleDeepLink(uri);
      }
    } catch (e) {
      print('Error handling initial URI: $e');
    }
  }

  void _handleIncomingLinks() {
    _deepLinkSubscription?.cancel();
    _deepLinkSubscription = _appLinks.uriLinkStream.listen(
      (Uri uri) {
        _handleDeepLink(uri);
      },
      onError: (error) {
        print('Error handling incoming links: $error');
      },
    );
  }

  void _handleDeepLink(Uri uri) {
    if (uri.scheme == 'vibehealth' && uri.host == 'vibe') {
      final typeParam = uri.queryParameters['type'];
      final valueParam = uri.queryParameters['value'];

      if (typeParam != null) {
        VibeType? vibeType;
        switch (typeParam) {
          case 'sleep':
            vibeType = VibeType.sleep;
            break;
          case 'mood':
            vibeType = VibeType.mood;
            break;
          default:
            vibeType = null;
        }

        if (vibeType != null) {
          int? initialValue;
          if (valueParam != null) {
            initialValue = int.tryParse(valueParam);
          }

          // Navigate to vibe entry screen
          // This will be handled after the app is fully initialized
          WidgetsBinding.instance.addPostFrameCallback((_) {
            Navigator.of(context).push(
              MaterialPageRoute(
                builder: (context) => VibeEntryScreen(
                  vibeType: vibeType!,
                  initialValue: initialValue,
                ),
              ),
            );
          });
        }
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Vibe Health',
      theme: ThemeData(
        colorScheme: ColorScheme.fromSeed(seedColor: Colors.teal),
        useMaterial3: true,
      ),
      initialRoute: '/',
      routes: {
        '/': (context) => const VibeDashboardScreen(),
        '/settings/notifications': (context) =>
            const NotificationSettingsScreen(),
      },
      onGenerateRoute: (settings) {
        if (settings.name == '/vibe/entry') {
          // Extract the arguments
          final args = settings.arguments as Map<String, dynamic>?;
          final typeString = args?['type'] as String?;

          VibeType vibeType;
          switch (typeString) {
            case 'sleep':
              vibeType = VibeType.sleep;
              break;
            case 'mood':
              vibeType = VibeType.mood;
              break;
            default:
              vibeType =
                  VibeType.mood; // Default to mood if type is not specified
          }

          return MaterialPageRoute(
            builder: (context) => VibeEntryScreen(
              vibeType: vibeType,
              initialValue: args?['value'] as int?,
              vibeId: args?['id'] as String?,
            ),
          );
        }
        return null;
      },
    );
  }
}

class MyHomePage extends StatefulWidget {
  const MyHomePage({super.key, required this.title});

  // This widget is the home page of your application. It is stateful, meaning
  // that it has a State object (defined below) that contains fields that affect
  // how it looks.

  // This class is the configuration for the state. It holds the values (in this
  // case the title) provided by the parent (in this case the App widget) and
  // used by the build method of the State. Fields in a Widget subclass are
  // always marked "final".

  final String title;

  @override
  State<MyHomePage> createState() => _MyHomePageState();
}

class _MyHomePageState extends State<MyHomePage> {
  int _counter = 0;

  void _incrementCounter() {
    setState(() {
      // This call to setState tells the Flutter framework that something has
      // changed in this State, which causes it to rerun the build method below
      // so that the display can reflect the updated values. If we changed
      // _counter without calling setState(), then the build method would not be
      // called again, and so nothing would appear to happen.
      _counter++;
    });
  }

  @override
  Widget build(BuildContext context) {
    // This method is rerun every time setState is called, for instance as done
    // by the _incrementCounter method above.
    //
    // The Flutter framework has been optimized to make rerunning build methods
    // fast, so that you can just rebuild anything that needs updating rather
    // than having to individually change instances of widgets.
    return Scaffold(
      appBar: AppBar(
        // TRY THIS: Try changing the color here to a specific color (to
        // Colors.amber, perhaps?) and trigger a hot reload to see the AppBar
        // change color while the other colors stay the same.
        backgroundColor: Theme.of(context).colorScheme.inversePrimary,
        // Here we take the value from the MyHomePage object that was created by
        // the App.build method, and use it to set our appbar title.
        title: Text(widget.title),
      ),
      body: Center(
        // Center is a layout widget. It takes a single child and positions it
        // in the middle of the parent.
        child: Column(
          // Column is also a layout widget. It takes a list of children and
          // arranges them vertically. By default, it sizes itself to fit its
          // children horizontally, and tries to be as tall as its parent.
          //
          // Column has various properties to control how it sizes itself and
          // how it positions its children. Here we use mainAxisAlignment to
          // center the children vertically; the main axis here is the vertical
          // axis because Columns are vertical (the cross axis would be
          // horizontal).
          //
          // TRY THIS: Invoke "debug painting" (choose the "Toggle Debug Paint"
          // action in the IDE, or press "p" in the console), to see the
          // wireframe for each widget.
          mainAxisAlignment: MainAxisAlignment.center,
          children: <Widget>[
            const Text('You have pushed the button this many times:'),
            Text(
              '$_counter',
              style: Theme.of(context).textTheme.headlineMedium,
            ),
          ],
        ),
      ),
      floatingActionButton: FloatingActionButton(
        onPressed: _incrementCounter,
        tooltip: 'Increment',
        child: const Icon(Icons.add),
      ), // This trailing comma makes auto-formatting nicer for build methods.
    );
  }
}
