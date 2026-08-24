import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../../core/validators/validators.dart';
import '../providers/auth_providers.dart';
import '../widgets/auth_widgets.dart';

class LoginPage extends ConsumerStatefulWidget {
  const LoginPage({super.key});
  @override
  ConsumerState<LoginPage> createState() => _LoginPageState();
}

class _LoginPageState extends ConsumerState<LoginPage> {
  final _formKey = GlobalKey<FormState>();
  final _usernameController = TextEditingController();

  @override
  void dispose() {
    _usernameController.dispose();
    super.dispose();
  }

  Future<void> _login() async {
    if (!_formKey.currentState!.validate()) return;
    await ref
        .read(authControllerProvider.notifier)
        .login(_usernameController.text.trim());
    // Navigation happens via the router redirect once auth state changes.
  }

  @override
  Widget build(BuildContext context) {
    final authState = ref.watch(authControllerProvider);

    return AuthFormScaffold(
      formKey: _formKey,
      controller: _usernameController,
      title: 'Welcome back',
      subtitle: 'Sign in securely using WebAuthn',
      fieldLabel: 'Workspace ID or Username',
      fieldHint: 'Ex. johndoe',
      buttonLabel: 'Continue with WebAuthn',
      loading: authState.loading,
      error: authState.error,
      validator: Validators.username,
      onSubmit: _login,
      footer: TextButton(
        onPressed: () {
          ref.read(authControllerProvider.notifier).clearError();
          context.go('/register');
        },
        child: const Text("Don't have an account? Register now"),
      ),
    );
  }
}
