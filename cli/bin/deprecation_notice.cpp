// Copyright (c) 2023 Proton AG
//
// This file is part of Proton Export Tool.
//
// Proton Mail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Proton Mail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Proton Export Tool.  If not, see <https://www.gnu.org/licenses/>.

#include <iostream>
#include <string>
#include <cstdlib>

void printDeprecationNotice() {
    std::cout << "\n";
    std::cout << "╔══════════════════════════════════════════════════════════════════════════════╗\n";
    std::cout << "║                              DEPRECATION NOTICE                             ║\n";
    std::cout << "╠══════════════════════════════════════════════════════════════════════════════╣\n";
    std::cout << "║                                                                              ║\n";
    std::cout << "║  The C++ CLI interface is deprecated and will be removed in a future        ║\n";
    std::cout << "║  version. Please use the new Go-based TUI interface instead.               ║\n";
    std::cout << "║                                                                              ║\n";
    std::cout << "║  New features:                                                               ║\n";
    std::cout << "║  • Interactive Terminal User Interface (TUI)                                ║\n";
    std::cout << "║  • Enhanced progress reporting                                               ║\n";
    std::cout << "║  • Better error handling and recovery                                       ║\n";
    std::cout << "║  • Improved user experience                                                  ║\n";
    std::cout << "║                                                                              ║\n";
    std::cout << "║  Migration:                                                                  ║\n";
    std::cout << "║  • Use 'proton-mail-export' instead of 'proton-mail-export-cli'           ║\n";
    std::cout << "║  • Add '--no-tui' flag to use CLI mode if needed                           ║\n";
    std::cout << "║  • All existing CLI arguments are supported                                 ║\n";
    std::cout << "║                                                                              ║\n";
    std::cout << "║  This notice can be suppressed with ET_SUPPRESS_DEPRECATION_WARNING=1      ║\n";
    std::cout << "║                                                                              ║\n";
    std::cout << "╚══════════════════════════════════════════════════════════════════════════════╝\n";
    std::cout << "\n";
    std::cout << "Continuing with C++ CLI in 5 seconds... (Press Ctrl+C to cancel)\n";
    
    // Wait 5 seconds unless suppressed
    if (std::getenv("ET_SUPPRESS_DEPRECATION_WARNING") == nullptr) {
        for (int i = 5; i > 0; i--) {
            std::cout << "\rStarting in " << i << " seconds...";
            std::cout.flush();
#ifdef _WIN32
            Sleep(1000);
#else
            sleep(1);
#endif
        }
        std::cout << "\rStarting now...                \n\n";
    }
}

bool shouldShowDeprecationNotice() {
    // Don't show if explicitly suppressed
    if (std::getenv("ET_SUPPRESS_DEPRECATION_WARNING") != nullptr) {
        return false;
    }
    
    // Don't show if running in non-interactive mode
    if (std::getenv("CI") != nullptr || std::getenv("AUTOMATED") != nullptr) {
        return false;
    }
    
    return true;
}