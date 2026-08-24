import 'package:flutter/material.dart';

/// Shared layout for login/register: icon, title, subtitle, one text field,
/// inline error, submit button, footer link.
class AuthFormScaffold extends StatelessWidget {
  final GlobalKey<FormState> formKey;
  final TextEditingController controller;
  final String title;
  final String subtitle;
  final String fieldLabel;
  final String fieldHint;
  final String buttonLabel;
  final bool loading;
  final String? error;
  final String? Function(String?)? validator;
  final VoidCallback onSubmit;
  final Widget footer;

  const AuthFormScaffold({
    super.key,
    required this.formKey,
    required this.controller,
    required this.title,
    required this.subtitle,
    required this.fieldLabel,
    required this.fieldHint,
    required this.buttonLabel,
    required this.loading,
    required this.error,
    required this.onSubmit,
    required this.footer,
    this.validator,
  });

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Center(
        child: ConstrainedBox(
          constraints: const BoxConstraints(maxWidth: 400),
          child: Padding(
            padding: const EdgeInsets.all(24),
            child: Form(
              key: formKey,
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  const Icon(Icons.fingerprint, size: 48),
                  const SizedBox(height: 24),
                  Text(
                    title,
                    style: Theme.of(context).textTheme.headlineSmall?.copyWith(
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                  const SizedBox(height: 8),
                  Text(
                    subtitle,
                    textAlign: TextAlign.center,
                    style: Theme.of(context).textTheme.bodyMedium,
                  ),
                  const SizedBox(height: 32),
                  TextFormField(
                    controller: controller,
                    enabled: !loading,
                    validator: validator,
                    decoration: InputDecoration(
                      labelText: fieldLabel,
                      hintText: fieldHint,
                      border: const OutlineInputBorder(),
                    ),
                  ),
                  if (error != null) ...[
                    const SizedBox(height: 16),
                    Text(
                      error!,
                      style: TextStyle(
                        color: Theme.of(context).colorScheme.error,
                      ),
                    ),
                  ],
                  const SizedBox(height: 20),
                  SizedBox(
                    width: double.infinity,
                    height: 48,
                    child: FilledButton(
                      onPressed: loading ? null : onSubmit,
                      child: loading
                          ? const SizedBox(
                              height: 20,
                              width: 20,
                              child: CircularProgressIndicator(strokeWidth: 2),
                            )
                          : Text(buttonLabel),
                    ),
                  ),
                  const SizedBox(height: 24),
                  footer,
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }
}
