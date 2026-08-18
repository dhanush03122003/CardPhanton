import 'package:flutter/material.dart';

import 'design_system/primary_scaffold.dart';

class PlaceholderPage extends StatelessWidget {
  const PlaceholderPage({
    required this.title,
    super.key,
  });

  final String title;

  @override
  Widget build(BuildContext context) {
    return PrimaryScaffold(
      title: title,
      child: Center(
        child: Text(
          '$title module shell',
          style: Theme.of(context).textTheme.titleMedium,
        ),
      ),
    );
  }
}
