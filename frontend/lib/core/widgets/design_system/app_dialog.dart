import 'package:flutter/material.dart';

class AppDialog extends StatelessWidget {
  const AppDialog({
    required this.title,
    required this.message,
    super.key,
    this.onConfirm,
    this.confirmLabel = 'OK',
  });

  final String title;
  final String message;
  final VoidCallback? onConfirm;
  final String confirmLabel;

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      title: Text(title),
      content: Text(message),
      actions: <Widget>[
        TextButton(
          onPressed: () {
            Navigator.of(context).pop();
            onConfirm?.call();
          },
          child: Text(confirmLabel),
        ),
      ],
    );
  }
}
