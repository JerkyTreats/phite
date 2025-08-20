// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'vibe_dto.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

VibeDto _$VibeDtoFromJson(Map<String, dynamic> json) => VibeDto(
  id: json['id'] as String?,
  userId: json['userId'] as String?,
  type: json['type'] as String,
  value: (json['value'] as num).toInt(),
  note: json['note'] as String?,
  ts: json['ts'] as String,
);

Map<String, dynamic> _$VibeDtoToJson(VibeDto instance) => <String, dynamic>{
  'id': instance.id,
  'userId': instance.userId,
  'type': instance.type,
  'value': instance.value,
  'note': instance.note,
  'ts': instance.ts,
};
