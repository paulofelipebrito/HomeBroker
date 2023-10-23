import Image from 'next/image'
import { Wallet } from './models';
import React from 'react';
import Link from 'next/link';

async function getWallets(): Promise<Wallet[]> {
  try{
    const response = await fetch(
      `http://host.docker.internal:3000/wallets`,
      {
        next: {
          tags: [`wallets`],
          //revalidate: isHomeBrokerClosed() ? 60 * 60 : 5,
          revalidate: 10,
        },
      },
    );
    return response.json();
  } catch (err) {
    console.error(err);
    return [];
  }
}

export default async function Home() {
  const wallets = await getWallets();
  console.log(wallets)

  return (
    <main className="flex min-h-screen items-start justify-start p-24 flex-wrap gap-28">
      {wallets?.map((el) => (
        <Link className='bg-white shadow-md w-60 h-56' key={el.id} href={`/${el.id}`}>
          <h2>{el.id}</h2>
          {el?.walletAssets?.map((asset) => (
            <React.Fragment key={asset.id}>
              <span>{asset.Asset.symbol}</span>
              <span>{asset.shares}</span>
            </React.Fragment>
          ))}
        </Link>
      ))}
    </main>
  )
}

