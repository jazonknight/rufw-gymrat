# GymRat

A personalized, file-based fitness data management engine built in Go. GymRat is designed to give you absolute control over your training data without relying on third-party cloud apps, walled gardens, or heavy database overhead.

## What is GymRat?

GymRat is a highly efficient, lightweight backend utility and CLI tool tailored for managing structured fitness data. It handles the end-to-end pipeline of your personal physical training data by structuring, modifying, and persisting your custom routines.

## Core Design

The architecture is built around clean, cloud-native Go philosophies:

* **File-Based Persistence:** It avoids the complexity of an external database engine, relying instead on flat-file persistence (like highly structured JSON or YAML) that easily tracks in a Git repository.
* **Strict Domain Modeling:** The core data layer leverages Go’s type system to handle complex, nested relationships:
  * **Plans:** High-level training blocks or macrocycles.
  * **Workouts:** Specific daily training sessions (such as lower body power or push days).
  * **Exercises:** Individual movements tracking metrics like sets, reps, weight, RPE (Rate of Perceived Exertion), and rest intervals.
* **Relational Lookups and Modularity:** The code is strictly modularized into independent packages. It uses relational lookup logic to tie individual exercise data back to global plans while maintaining clean, separate domain models.
* **Local and Shared Packages:** The codebase is designed to easily share these models across multiple applications (such as a CLI tool, local scripts, or potential web frontends) using local Go workspaces or GitHub repositories.

## Why Build It This Way?

### Digital Minimalism and Data Sovereignty
Commercial fitness apps are packed with bloat, ads, social feeds, and changing monetization models. GymRat provides maximum utility with zero digital noise. Because the data lives in flat files, you own it permanently, it remains entirely private, and it can be parsed or migrated effortlessly.

### Performance and Simplicity
By choosing Go and file-based data structures over an enterprise relational database or a heavy ORM (Object-Relational Mapping), the tool achieves near-instantaneous execution times. It aligns perfectly with a systems-engineering approach: fast, predictable, and simple to maintain.

### Extensibility and the RupertFrameworks Connection
Because the codebase is modular, it acts as a perfect practical sandbox for testing high-performance Go utilities, clean architecture paradigms, and shared module strategies. The design ensures that if you want to build a frontend interface or automate data syncing later, the foundational core is already built to scale.