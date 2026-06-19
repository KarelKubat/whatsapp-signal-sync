#!/usr/bin/env perl

use strict;
use warnings;
use Fcntl qw(:flock);

sub cmd_loglimit {
    die("Usage: loglimit FILE SIZE HISTORYFILES\n")
      if ( $#ARGV != 2 );

    my ( $file, $size, $historyfiles ) = @ARGV;

    $size = int($size);
    die("Bad size '$size'\n") if ( $size <= 1 );

    $historyfiles = int($historyfiles);
    die("Bad history files number '$historyfiles'\n")
      if ( $historyfiles < 0 );

    my $of;
    open( $of, ">>$file" ) or die("Cannot write $file: $!\n");
    select($of);
    $|++;
    select(STDOUT);

    open( my $lockf, '/dev/null' )
      or die("Cannot open lock file /dev/null: $!\n");

    while ( my $line = <STDIN> ) {
        flock( $lockf, LOCK_EX );
        if ( !print $of ($line) ) {
            warn("Cannot append $file: $!\n");
            open( $of, ">$file" ) or die("Cannot rewrite $file: $!\n");
            select($of);
            $|++;
            select(STDOUT);
            next;
        }
        if ( ( stat($of) )[7] > $size ) {

            # Close stream, rotate old files
            close($of);
            for ( my $nr = $historyfiles ; $nr > 1 ; $nr-- ) {
                my $lower = $nr - 1;
                unlink("$file.$nr");
                rename( "$file.$lower", "$file.$nr" );
            }
            rename( $file, "$file.1" );

            # Open new stream
            open( $of, ">$file" ) or die("Cannot write new $file: $!\n");
            select($of);
            $|++;
            select(STDOUT);
        }
        flock( $lockf, LOCK_UN );
    }
}

# Main
cmd_loglimit(@ARGV);
