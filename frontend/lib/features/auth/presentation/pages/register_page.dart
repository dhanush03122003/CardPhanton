import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../../core/validators/validators.dart';
import '../providers/auth_providers.dart';
import '../widgets/auth_widgets.dart';

class RegisterPage extends ConsumerStatefulWidget {
  const RegisterPage({super.key});
  @override
  ConsumerState<RegisterPage> createState() => _RegisterPageState();
}

class _RegisterPageState extends ConsumerState<RegisterPage> {
  final _formKey = GlobalKey<FormState>();
  final _usernameController = TextEditingController();

  @override
  void dispose() {
    _usernameController.dispose();
    super.dispose();
  }

  Future<void> _register() async {
    if (!_formKey.currentState!.validate()) return;
    await ref
        .read(authControllerProvider.notifier)
        .register(_usernameController.text.trim());
  }

  @override
  Widget build(BuildContext context) {
    final authState = ref.watch(authControllerProvider);

    return AuthFormScaffold(
      formKey: _formKey,
      controller: _usernameController,
      title: 'Create an account',
      subtitle: 'Register securely using WebAuthn',
      fieldLabel: 'Choose a Username',
      fieldHint: 'Ex. johndoe',
      buttonLabel: 'Register with WebAuthn',
      loading: authState.loading,
      error: authState.error,
      validator: Validators.username,
      onSubmit: _register,
      footer: TextButton(
        onPressed: () {
          ref.read(authControllerProvider.notifier).clearError();
          context.go('/login');
        },
        child: const Text('Already have an account? Sign in'),
      ),
    );
  }
}
