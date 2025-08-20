import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:intl/intl.dart';

import '../../controllers/vibe_controller.dart';
import '../../domain/vibe.dart';
import '../../domain/vibe_type.dart';
import '../widgets/app_drawer.dart';
import '../widgets/vibe_list_item.dart';
import 'vibe_entry_screen.dart';

/// Screen for displaying vibe history and summary
class VibeDashboardScreen extends ConsumerStatefulWidget {
  const VibeDashboardScreen({super.key});

  @override
  ConsumerState<VibeDashboardScreen> createState() => _VibeDashboardScreenState();
}

class _VibeDashboardScreenState extends ConsumerState<VibeDashboardScreen> {
  DateTime _selectedDate = DateTime.now();
  bool _isLoading = false;

  @override
  void initState() {
    super.initState();
    _loadVibes();
  }

  Future<void> _loadVibes() async {
    setState(() {
      _isLoading = true;
    });

    try {
      // Load vibes for the selected date
      final startOfDay = DateTime(_selectedDate.year, _selectedDate.month, _selectedDate.day);
      final endOfDay = DateTime(_selectedDate.year, _selectedDate.month, _selectedDate.day, 23, 59, 59);
      
      await ref.read(vibeControllerProvider.notifier).loadVibes(
        from: startOfDay,
        to: endOfDay,
      );
    } finally {
      if (mounted) {
        setState(() {
          _isLoading = false;
        });
      }
    }
  }

  Future<void> _selectDate(BuildContext context) async {
    final DateTime? picked = await showDatePicker(
      context: context,
      initialDate: _selectedDate,
      firstDate: DateTime(2020),
      lastDate: DateTime.now(),
    );
    
    if (picked != null && picked != _selectedDate) {
      setState(() {
        _selectedDate = picked;
      });
      _loadVibes();
    }
  }

  void _navigateToEntryScreen(VibeType type) {
    Navigator.of(context).push(
      MaterialPageRoute(
        builder: (context) => VibeEntryScreen(vibeType: type),
      ),
    ).then((_) => _loadVibes());
  }

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(vibeControllerProvider);
    final sleepVibes = state.vibes.where((v) => v.type == VibeType.sleep).toList();
    final moodVibes = state.vibes.where((v) => v.type == VibeType.mood).toList();

    return Scaffold(
      appBar: AppBar(
        title: const Text('Vibe Dashboard'),
        actions: [
          IconButton(
            icon: const Icon(Icons.calendar_today),
            onPressed: () => _selectDate(context),
          ),
        ],
      ),
      drawer: const AppDrawer(),
      body: RefreshIndicator(
        onRefresh: _loadVibes,
        child: _isLoading || state.isLoading
            ? const Center(child: CircularProgressIndicator())
            : state.error != null
                ? _buildErrorView(state.error!)
                : _buildDashboard(sleepVibes, moodVibes),
      ),
      floatingActionButton: FloatingActionButton(
        onPressed: () {
          showModalBottomSheet(
            context: context,
            builder: (context) => _buildEntryOptions(),
          );
        },
        child: const Icon(Icons.add),
      ),
    );
  }

  Widget _buildErrorView(dynamic error) {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          const Icon(
            Icons.error_outline,
            color: Colors.red,
            size: 60,
          ),
          const SizedBox(height: 16),
          Text(
            'Error: ${error.toString()}',
            textAlign: TextAlign.center,
            style: const TextStyle(fontSize: 16),
          ),
          const SizedBox(height: 24),
          ElevatedButton(
            onPressed: () {
              ref.read(vibeControllerProvider.notifier).clearError();
              _loadVibes();
            },
            child: const Text('Try Again'),
          ),
        ],
      ),
    );
  }

  Widget _buildDashboard(List<Vibe> sleepVibes, List<Vibe> moodVibes) {
    final dateFormat = DateFormat('EEEE, MMMM d, y');
    
    return SingleChildScrollView(
      padding: const EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            dateFormat.format(_selectedDate),
            style: Theme.of(context).textTheme.headlineSmall,
          ),
          const SizedBox(height: 24),
          
          // Sleep section
          _buildVibeSection(
            title: 'Sleep',
            vibeType: VibeType.sleep,
            vibes: sleepVibes,
          ),
          const SizedBox(height: 24),
          
          // Mood section
          _buildVibeSection(
            title: 'Mood',
            vibeType: VibeType.mood,
            vibes: moodVibes,
          ),
        ],
      ),
    );
  }

  Widget _buildVibeSection({
    required String title,
    required VibeType vibeType,
    required List<Vibe> vibes,
  }) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            Text(
              title,
              style: Theme.of(context).textTheme.titleLarge,
            ),
            TextButton.icon(
              onPressed: () => _navigateToEntryScreen(vibeType),
              icon: const Icon(Icons.add),
              label: const Text('Add'),
            ),
          ],
        ),
        const SizedBox(height: 8),
        _buildVibeList(title, vibes),
      ],
    );
  }

  Widget _buildVibeList(String title, List<Vibe> vibes) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Padding(
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
          child: Text(
            title,
            style: const TextStyle(
              fontSize: 18,
              fontWeight: FontWeight.bold,
            ),
          ),
        ),
        vibes.isEmpty
            ? const Padding(
                padding: EdgeInsets.all(16),
                child: Center(
                  child: Text('No vibes recorded yet'),
                ),
              )
            : ListView.builder(
                shrinkWrap: true,
                physics: const NeverScrollableScrollPhysics(),
                padding: const EdgeInsets.symmetric(horizontal: 16),
                itemCount: vibes.length,
                itemBuilder: (context, index) {
                  final vibe = vibes[index];
                  return VibeListItem(
                    vibe: vibe,
                    onTap: () {
                      // Navigate to edit screen
                      Navigator.pushNamed(
                        context,
                        '/vibe/entry',
                        arguments: {
                          'type': vibe.type.name,
                          'value': vibe.value,
                          'id': vibe.id,
                        },
                      );
                    },
                  );
                },
              ),
      ],
    );
  }

  Widget _buildEntryOptions() {
    return Container(
      padding: const EdgeInsets.all(16),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Text(
            'Record a new vibe',
            style: Theme.of(context).textTheme.titleLarge,
          ),
          const SizedBox(height: 16),
          ListTile(
            leading: const Icon(Icons.bedtime),
            title: const Text('Sleep Check'),
            onTap: () {
              Navigator.of(context).pop();
              _navigateToEntryScreen(VibeType.sleep);
            },
          ),
          ListTile(
            leading: const Icon(Icons.mood),
            title: const Text('Mood Check'),
            onTap: () {
              Navigator.of(context).pop();
              _navigateToEntryScreen(VibeType.mood);
            },
          ),
        ],
      ),
    );
  }
}
