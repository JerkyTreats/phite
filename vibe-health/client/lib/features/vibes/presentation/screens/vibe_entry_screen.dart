import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../controllers/vibe_controller.dart';
import '../../domain/vibe.dart';
import '../../domain/vibe_type.dart';
import '../widgets/app_drawer.dart';

/// Screen for entering or editing a vibe
class VibeEntryScreen extends ConsumerStatefulWidget {
  final VibeType vibeType;
  final int? initialValue;
  final String? vibeId;

  const VibeEntryScreen({
    super.key,
    required this.vibeType,
    this.initialValue,
    this.vibeId,
  });

  @override
  ConsumerState<VibeEntryScreen> createState() => _VibeEntryScreenState();
}

class _VibeEntryScreenState extends ConsumerState<VibeEntryScreen> {
  late int _selectedValue;
  final TextEditingController _noteController = TextEditingController();
  bool _isSubmitting = false;
  bool _isEditing = false;
  Vibe? _existingVibe;

  @override
  void initState() {
    super.initState();
    _selectedValue = widget.initialValue ?? 3;
    _isEditing = widget.vibeId != null;
    
    if (_isEditing) {
      _loadExistingVibe();
    }
  }
  
  Future<void> _loadExistingVibe() async {
    final state = ref.read(vibeControllerProvider);
    final vibe = state.vibes.firstWhere(
      (v) => v.id == widget.vibeId,
      orElse: () => Vibe(
        id: 'temp',
        userId: 'current_user',
        type: widget.vibeType,
        value: _selectedValue,
        ts: DateTime.now().toUtc().toIso8601String(),
      ),
    );
    
    setState(() {
      _existingVibe = vibe;
      _selectedValue = vibe.value;
      if (vibe.note != null) {
        _noteController.text = vibe.note!;
      }
    });
  }

  @override
  void dispose() {
    _noteController.dispose();
    super.dispose();
  }

  Future<void> _submitVibe() async {
    if (_isSubmitting) return;

    setState(() {
      _isSubmitting = true;
    });

    try {
      final note = _noteController.text.isNotEmpty ? _noteController.text : null;
      
      if (_isEditing && _existingVibe != null) {
        // Update existing vibe
        final updatedVibe = _existingVibe!.copyWith(
          value: _selectedValue,
          note: note,
        );
        
        await ref.read(vibeControllerProvider.notifier).updateVibe(updatedVibe);
        
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(
              content: Text('${widget.vibeType.displayName} updated successfully!'),
              backgroundColor: Colors.green,
            ),
          );
          Navigator.of(context).pop();
        }
      } else {
        // Create new vibe
        await ref.read(vibeControllerProvider.notifier).createVibe(
              widget.vibeType,
              _selectedValue,
              note: note,
            );

        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(
              content: Text('${widget.vibeType.displayName} recorded successfully!'),
              backgroundColor: Colors.green,
            ),
          );
          Navigator.of(context).pop();
        }
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(
              _isEditing
                  ? 'Failed to update ${widget.vibeType.displayName}: $e'
                  : 'Failed to record ${widget.vibeType.displayName}: $e',
            ),
            backgroundColor: Colors.red,
          ),
        );
      }
    } finally {
      if (mounted) {
        setState(() {
          _isSubmitting = false;
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(vibeControllerProvider);

    return Scaffold(
      appBar: AppBar(
        title: Text(_isEditing 
            ? 'Edit ${widget.vibeType.displayName}'
            : '${widget.vibeType.displayName} Check'),
      ),
      drawer: const AppDrawer(),
      body: state.error != null
          ? _buildErrorView(state.error!)
          : _buildEntryForm(),
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
            },
            child: const Text('Try Again'),
          ),
        ],
      ),
    );
  }

  Widget _buildEntryForm() {
    return SingleChildScrollView(
      padding: const EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Text(
            widget.vibeType.promptQuestion,
            style: Theme.of(context).textTheme.headlineSmall,
            textAlign: TextAlign.center,
          ),
          const SizedBox(height: 32),
          _buildEmojiSelector(),
          const SizedBox(height: 32),
          _buildNoteField(),
          const SizedBox(height: 24),
          ElevatedButton(
            onPressed: _isSubmitting ? null : _submitVibe,
            style: ElevatedButton.styleFrom(
              padding: const EdgeInsets.symmetric(vertical: 16),
            ),
            child: _isSubmitting
                ? const SizedBox(
                    height: 24,
                    width: 24,
                    child: CircularProgressIndicator(
                      strokeWidth: 2,
                    ),
                  )
                : Text(_isEditing ? 'Update' : 'Submit'),
          ),
        ],
      ),
    );
  }

  Widget _buildEmojiSelector() {
    return Row(
      mainAxisAlignment: MainAxisAlignment.spaceEvenly,
      children: [
        _buildEmojiButton('😡', 1),
        _buildEmojiButton('🙁', 2),
        _buildEmojiButton('😐', 3),
        _buildEmojiButton('🙂', 4),
        _buildEmojiButton('😁', 5),
      ],
    );
  }

  Widget _buildEmojiButton(String emoji, int value) {
    final isSelected = _selectedValue == value;
    return GestureDetector(
      onTap: () {
        setState(() {
          _selectedValue = value;
        });
      },
      child: Container(
        width: 56,
        height: 56,
        decoration: BoxDecoration(
          shape: BoxShape.circle,
          color: isSelected ? Theme.of(context).primaryColor.withOpacity(0.2) : null,
          border: isSelected
              ? Border.all(color: Theme.of(context).primaryColor, width: 2)
              : null,
        ),
        child: Center(
          child: Text(
            emoji,
            style: const TextStyle(fontSize: 32),
          ),
        ),
      ),
    );
  }

  Widget _buildNoteField() {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          'Add a note (optional):',
          style: Theme.of(context).textTheme.titleMedium,
        ),
        const SizedBox(height: 8),
        TextField(
          controller: _noteController,
          maxLines: 3,
          decoration: const InputDecoration(
            hintText: 'How are you feeling?',
            border: OutlineInputBorder(),
          ),
        ),
      ],
    );
  }
}
