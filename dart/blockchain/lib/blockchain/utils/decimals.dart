String rescaleDecimalString(String value, int fromDecimals, int toDecimals) {
  // Normalize decimal separator to dot
  value = value.replaceAll(',', '.');

  // Split into integer and fractional parts
  final parts = value.split('.');
  // Se tiver mais de um separador decimal, é inválido e deve falhar
  if (parts.length > 2) {
    throw FormatException('Invalid numeric input');
  }

  String intPart = parts[0];
  String fracPart = parts.length > 1 ? parts[1] : '';

  // Validate all characters are numeric
  final allChars = intPart + fracPart;
  if (!RegExp(r'^\d*$').hasMatch(allChars)) {
    throw FormatException('Invalid numeric input');
  }

  // Trim leading zeros in int part
  intPart = intPart.replaceFirst(RegExp(r'^0+'), '');
  if (intPart.isEmpty) intPart = '0';

  // Right-pad or truncate fractional part to match fromDecimals
  if (fracPart.length < fromDecimals) {
    fracPart = fracPart.padRight(fromDecimals, '0');
  } else {
    fracPart = fracPart.substring(0, fromDecimals);
  }

  // Combine parts
  String combined = intPart + fracPart;
  combined = combined.replaceFirst(RegExp(r'^0+'), '');
  if (combined.isEmpty) combined = '0';

  // Rescale to toDecimals
  final diff = toDecimals - fromDecimals;
  if (diff == 0) {
    return combined;
  } else if (diff > 0) {
    return combined + '0' * diff;
  } else {
    if (combined.length <= -diff) {
      return '0'; // underflow
    }
    return combined.substring(0, combined.length + diff);
  }
}
