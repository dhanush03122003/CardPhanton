import 'failures.dart';

sealed class Result<T> {
  const Result();

  bool get isSuccess => this is Success<T>;
  bool get isFailure => this is Error<T>;
}

class Success<T> extends Result<T> {
  const Success(this.data);

  final T data;
}

class Error<T> extends Result<T> {
  const Error(this.failure);

  final Failure failure;
}
