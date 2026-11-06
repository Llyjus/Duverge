package com.duverge.userdata.error;

import org.hibernate.exception.ConstraintViolationException;

/** Identifies named database constraint violations without guessing from the operation. */
public final class DatabaseConstraintClassifier {
    private DatabaseConstraintClassifier() {
    }

    public static boolean violates(Throwable error, String expectedConstraint) {
        for (Throwable cause = error; cause != null; cause = cause.getCause()) {
            if (cause instanceof ConstraintViolationException violation) {
                String actualConstraint = violation.getConstraintName();
                // Drivers and dialects may include schema/index details around
                // the declared constraint name (H2 does this, for example).
                if (actualConstraint != null && java.util.regex.Pattern.compile(
                        "(?i)(?:^|[^a-z0-9_])" + java.util.regex.Pattern.quote(expectedConstraint)
                                + "(?:$|[^a-z0-9_])")
                        .matcher(actualConstraint).find()) {
                    return true;
                }
            }
        }
        return false;
    }
}
