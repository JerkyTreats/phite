import 'package:flutter/material.dart';
import 'package:intl/intl.dart';

import '../../domain/vibe.dart';
import '../../domain/vibe_type.dart';

/// Widget to display a single vibe item in a list
class VibeListItem extends StatelessWidget {
  final Vibe vibe;
  final VoidCallback? onTap;

  const VibeListItem({
    super.key,
    required this.vibe,
    this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return Card(
      margin: const EdgeInsets.symmetric(vertical: 8, horizontal: 4),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(8),
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Row(
                    children: [
                      _buildVibeIcon(),
                      const SizedBox(width: 12),
                      Text(
                        vibe.type.displayName,
                        style: Theme.of(context).textTheme.titleMedium,
                      ),
                    ],
                  ),
                  Text(
                    _formatDate(vibe.ts),
                    style: Theme.of(context).textTheme.bodySmall,
                  ),
                ],
              ),
              const SizedBox(height: 12),
              Row(
                children: [
                  _buildVibeValue(context),
                  const SizedBox(width: 16),
                  Expanded(
                    child: Text(
                      _getVibeDescription(),
                      style: Theme.of(context).textTheme.bodyMedium,
                    ),
                  ),
                ],
              ),
              if (vibe.note != null && vibe.note!.isNotEmpty) ...[
                const SizedBox(height: 8),
                Text(
                  vibe.note!,
                  style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                    fontStyle: FontStyle.italic,
                    color: Colors.grey[700],
                  ),
                ),
              ],
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildVibeIcon() {
    IconData iconData;
    Color iconColor;

    switch (vibe.type) {
      case VibeType.sleep:
        iconData = Icons.bedtime;
        iconColor = Colors.indigo;
        break;
      case VibeType.mood:
        iconData = Icons.mood;
        iconColor = Colors.amber;
        break;
    }

    return Container(
      padding: const EdgeInsets.all(8),
      decoration: BoxDecoration(
        color: iconColor.withOpacity(0.1),
        borderRadius: BorderRadius.circular(8),
      ),
      child: Icon(
        iconData,
        color: iconColor,
      ),
    );
  }

  Widget _buildVibeValue(BuildContext context) {
    final value = vibe.value;
    Color valueColor;
    String valueText;

    if (vibe.type == VibeType.sleep) {
      if (value >= 8) {
        valueColor = Colors.green;
        valueText = 'Great';
      } else if (value >= 6) {
        valueColor = Colors.amber;
        valueText = 'Good';
      } else {
        valueColor = Colors.red;
        valueText = 'Poor';
      }
    } else { // Mood
      if (value >= 8) {
        valueColor = Colors.green;
        valueText = 'Great';
      } else if (value >= 5) {
        valueColor = Colors.amber;
        valueText = 'Good';
      } else if (value >= 3) {
        valueColor = Colors.orange;
        valueText = 'Okay';
      } else {
        valueColor = Colors.red;
        valueText = 'Poor';
      }
    }

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
      decoration: BoxDecoration(
        color: valueColor.withOpacity(0.1),
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: valueColor.withOpacity(0.3)),
      ),
      child: Row(
        children: [
          Text(
            '$value',
            style: Theme.of(context).textTheme.titleMedium?.copyWith(
              color: valueColor,
              fontWeight: FontWeight.bold,
            ),
          ),
          const SizedBox(width: 4),
          Text(
            '/ 10',
            style: Theme.of(context).textTheme.bodySmall?.copyWith(
              color: valueColor,
            ),
          ),
          const SizedBox(width: 8),
          Text(
            valueText,
            style: Theme.of(context).textTheme.bodySmall?.copyWith(
              color: valueColor,
              fontWeight: FontWeight.w500,
            ),
          ),
        ],
      ),
    );
  }

  String _formatDate(String timestamp) {
    final date = DateTime.parse(timestamp);
    return DateFormat('MMM d, h:mm a').format(date);
  }

  String _getVibeDescription() {
    if (vibe.type == VibeType.sleep) {
      if (vibe.value >= 8) {
        return 'Excellent sleep quality';
      } else if (vibe.value >= 6) {
        return 'Good sleep quality';
      } else if (vibe.value >= 4) {
        return 'Fair sleep quality';
      } else {
        return 'Poor sleep quality';
      }
    } else { // Mood
      if (vibe.value >= 8) {
        return 'Feeling great';
      } else if (vibe.value >= 6) {
        return 'Feeling good';
      } else if (vibe.value >= 4) {
        return 'Feeling okay';
      } else if (vibe.value >= 2) {
        return 'Feeling down';
      } else {
        return 'Feeling very low';
      }
    }
  }
}
